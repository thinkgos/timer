package timer

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/thinkgos/timer/delayqueue"
)

const (
	// DefaultTickMs default tick milliseconds.
	DefaultTickMs = 1
	// DefaultWheelSize default wheel size.
	DefaultWheelSize = 128
)

// ErrClosed is returned when the timer is closed.
var ErrClosed = errors.New("timer: use of closed timer")

// ErrInvalidDelay is returned when the task's delay is not greater than 0.
var ErrInvalidDelay = errors.New("timer: task delay must be greater than 0")

// goroutinePool is a reusable go pool.
var goroutinePool = goroutine{}

// GoPool goroutine pool, used to run the job of an expired task.
//
// NOTE: `Go` MUST NOT block. `Timer.addTaskEntry` submits the job while holding
// `Timer.rw`, and the job's re-schedule path calls `Timer.AddTask`, which needs
// `Timer.rw` too. So a bounded pool that blocks when full deadlocks the timer:
// the submit never returns, so the lock is never released, so no running job can
// ever complete and free a slot. An implementation backed by a bounded pool must
// fall back to a bare `go f()` on submit failure, as `wrapperAnts` does.
type GoPool interface {
	// Go run f. It must return without waiting for f, and must never block.
	Go(f func())
}

// DerefTask a container hold task
type DerefTask interface {
	DerefTask() *Task
}

// TaskContainer DerefTask's alias.
type TaskContainer = DerefTask

type goroutine struct{}

// Go implements GoPool interface.
func (goroutine) Go(f func()) {
	go f()
}

// Option `Timer` custom options.
type Option func(*Timer)

// WithTickMs set basic time tick milliseconds.
func WithTickMs(tickMs int64) Option {
	return func(t *Timer) {
		t.tickMs = tickMs
	}
}

// WithWheelSize set wheel size.
func WithWheelSize(size int) Option {
	return func(t *Timer) {
		t.wheelSize = NextPowOf2(size)
		t.wheelMask = t.wheelSize - 1
	}
}

// WithGoPool set goroutine pool, `p` must never block on submit. see [GoPool].
func WithGoPool(p GoPool) Option {
	return func(t *Timer) {
		t.goPool = p
	}
}

// Timer is a timer
type Timer struct {
	tickMs      int64                          // basic time span, unit is milliseconds.
	wheelSize   int                            // wheel size, the power of 2
	wheelMask   int                            // wheel mask
	taskCounter atomic.Int64                   // the total number of tasks.
	delayQueue  *delayqueue.DelayQueue[*Spoke] // delay queue, the priority queue use spoke's expiration time as `cmp`.
	goPool      GoPool                         // goroutine pool
	waitGroup   sync.WaitGroup                 // ensure the goroutine has finished.
	rw          sync.RWMutex                   // protects following fields.
	wheel       *TimingWheel                   // timing wheel, concurrent add task(read-lock) and advance clock only one(write-lock).
	quit        chan struct{}                  // of chan struct{}, created when first start.
	closed      bool                           // true if closed.
}

// NewTimer new timer instance. default tick is 1 milliseconds, wheel size is 128.
func NewTimer(opts ...Option) *Timer {
	t := &Timer{
		tickMs:      DefaultTickMs,
		wheelSize:   DefaultWheelSize,
		wheelMask:   DefaultWheelSize - 1,
		taskCounter: atomic.Int64{},
		delayQueue:  delayqueue.NewDelayQueue(CompareSpoke),
		goPool:      goroutinePool,
		quit:        nil,
		closed:      true,
	}
	for _, opt := range opts {
		opt(t)
	}
	if t.tickMs <= 0 {
		panic("timer: tick must be greater than or equal to 1ms")
	}
	if t.wheelSize <= 0 {
		panic("timer: wheel size must be greater than 0")
	}
	if t.goPool == nil {
		t.goPool = goroutinePool
	}
	t.wheel = newTimingWheel(t, t.tickMs, time.Now().UnixMilli())
	return t
}

// TickMs return basic time tick milliseconds.
func (t *Timer) TickMs() int64 { return t.tickMs }

// WheelSize return the wheel size.
func (t *Timer) WheelSize() int { return t.wheelSize }

// WheelMask return the wheel mask.
func (t *Timer) WheelMask() int { return t.wheelMask }

// TaskCounter return the total number of tasks.
func (t *Timer) TaskCounter() int64 { return t.taskCounter.Load() }

// AfterFunc adds a function to the timer.
// `d` must be greater than 0, otherwise it returns `ErrInvalidDelay`.
func (t *Timer) AfterFunc(d time.Duration, f func()) (*Task, error) {
	task := NewTask(d).WithJobFunc(f)
	err := t.AddTask(task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// AddTask adds a task to the timer.
// The task's delay must be greater than 0, otherwise it returns `ErrInvalidDelay`.
// A `Schedule` that has no next run reports a delay of -1, so re-adding an exhausted
// task reports `ErrInvalidDelay` rather than being silently dropped.
func (t *Timer) AddTask(task *Task) error {
	if task.Delay() <= 0 {
		return ErrInvalidDelay
	}
	t.rw.RLock()
	defer t.rw.RUnlock()
	if t.closed {
		return ErrClosed
	}
	t.addTaskEntry(newTaskEntry(task))
	return nil
}

// AddDerefTask adds a task from DerefTask to the timer.
func (t *Timer) AddDerefTask(tc DerefTask) error {
	return t.AddTask(tc.DerefTask())
}

// Started have started or not.
func (t *Timer) Started() bool {
	t.rw.RLock()
	defer t.rw.RUnlock()
	return !t.closed
}

// Start the timer.
func (t *Timer) Start() {
	t.rw.Lock()
	defer t.rw.Unlock()
	if t.closed {
		t.closed = false
		t.quit = make(chan struct{})
		t.waitGroup.Add(1)
		go func() {
			defer t.waitGroup.Done()
			for {
				spoke, exit := t.delayQueue.Take(t.quit)
				if exit {
					break
				}
				t.rw.Lock()
				for exist := true; exist; spoke, exist = t.delayQueue.Poll() {
					t.wheel.advanceClock(spoke.GetExpiration())
					spoke.Flush(t.addTaskEntry) // reinsert task entry to the timer
				}
				t.rw.Unlock()
			}
		}()
	}
}

// Stop the timer, graceful shutdown waiting the goroutine until it's stopped.
func (t *Timer) Stop() {
	t.rw.Lock()
	if !t.closed {
		close(t.quit)
		t.closed = true
	}
	t.rw.Unlock()

	t.waitGroup.Wait() // wait outside the lock to avoid deadlock with Start() goroutine
}

func (t *Timer) addToDelayQueue(spoke *Spoke) {
	t.delayQueue.Add(spoke)
}

// NOTE: should be call when `Timer.rw` lock.
func (t *Timer) addTaskEntry(te *taskEntry) {
	// if success, we do not need deal the task entry, because it has be added to the timing wheel.
	// if cancelled, we ignore the task entry.
	// if already expired, we run the task job.
	if t.wheel.add(te) == Result_AlreadyExpired {
		task := te.task
		t.goPool.Go(func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Fprintf(os.Stderr, "timer: Recovered from panic: %v\n", err)
				}
			}()
			defer func() {
				if delay := te.task.job.NextDelay(); delay > 0 {
					te.task.SetDelay(delay)
					_ = t.AddTask(task)
				}
			}()
			task.Run()
		})
	}
}
