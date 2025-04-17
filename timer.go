package timer

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/thinkgos/timer/delayqueue"
)

const (
	// DefaultTickMs default tick milliseconds.
	DefaultTickMs = 1
	// DefaultWheelSize default wheel size.
	DefaultWheelSize = 512
)

// ErrClosed is returned when the timer is closed.
var ErrClosed = errors.New("timer: use of closed timer")

// goroutinePool is a reusable go pool.
var goroutinePool = goroutine{}

// GoPool goroutine pool.
type GoPool interface {
	Go(f func())
}

// TaskContainer a container hold task
type TaskContainer interface {
	DerefTask() *Task
}

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

// WithGoPool set goroutine pool.
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
	taskCounter atomic.Int64                   // task total count
	delayQueue  *delayqueue.DelayQueue[*Spoke] // delay queue
	goPool      GoPool                         // goroutine pool
	waitGroup   sync.WaitGroup                 // ensure the goroutine has finished.
	rw          sync.RWMutex                   // protects following fields.
	wheel       *TimingWheel                   // timing wheel, concurrent add task(read-lock) and advance clock only one(write-lock).
	quit        chan struct{}                  // of chan struct{}, created when first start.
	closed      bool                           // true if closed.
}

// NewTimer new timer instance. default tick is 1 milliseconds, wheel size is 512.
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
func (t *Timer) AfterFunc(d time.Duration, f func()) (*Task, error) {
	task := NewTask(d).WithJobFunc(f)
	err := t.AddTask(task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// AddTask adds a task to the timer.
func (t *Timer) AddTask(task *Task) error {
	t.rw.RLock()
	defer t.rw.RUnlock()
	if t.closed {
		return ErrClosed
	}
	t.addTaskEntry(newTaskEntry(task))
	return nil
}

// AddDerefTask adds a task from TaskContainer to the timer.
func (t *Timer) AddDerefTask(tc TaskContainer) error {
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
				for exist := true; exist; spoke, exist = t.delayQueue.Poll() {
					t.advanceWheelClock(spoke.GetExpiration())
					t.flushSpoke(spoke)
				}
			}
		}()
	}
}

// Stop the timer, graceful shutdown waiting the goroutine until it's stopped.
func (t *Timer) Stop() {
	t.rw.Lock()
	defer t.rw.Unlock()
	if !t.closed {
		close(t.quit)
		t.waitGroup.Wait() // Ensure the goroutine has finished
		t.closed = true
	}
}

func (t *Timer) advanceWheelClock(expiration int64) {
	t.rw.Lock()
	defer t.rw.Unlock()
	t.wheel.advanceClock(expiration)
}

func (t *Timer) flushSpoke(spoke *Spoke) {
	t.rw.RLock()
	defer t.rw.RUnlock()
	spoke.Flush(t.reinsert)
}

func (t *Timer) addSpokeToDelayQueue(spoke *Spoke) {
	t.delayQueue.Add(spoke)
}

func (t *Timer) addTaskEntry(te *taskEntry) {
	// if success, we do not need deal the task entry, because it has be added to the timing wheel.
	// if cancelled cancelled, we ignore the task entry.
	// if already expired, we run the task job.
	if t.wheel.add(te) == Result_AlreadyExpired {
		t.goPool.Go(te.task.Run)
	}
}

func (t *Timer) reinsert(te *taskEntry) {
	t.addTaskEntry(te)
}
