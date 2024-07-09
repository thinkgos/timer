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
	// DefaultTimeUnit default time unit is milliseconds.
	DefaultTimeUnit = time.Millisecond
)

var (
	// ErrClosed is returned when the timer is closed.
	ErrClosed = errors.New("timer: use of closed timer")
	// closedchan is a reusable closed channel.
	closedchan = make(chan struct{})
	// goroutinePool is a reusable go pool.
	goroutinePool = goroutine{}
)

func init() {
	close(closedchan)
}

// GoPool goroutine pool.
type GoPool interface {
	Go(f func())
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
	tickMs      int64                          // 基本时间跨度, 单位ms
	wheelSize   int                            // 轮的大小, 2的n次方
	wheelMask   int                            // 轮的掩码
	taskCounter atomic.Int64                   // 任务总数
	delayQueue  *delayqueue.DelayQueue[*Spoke] // 延迟队列
	goPool      GoPool                         // 协程池
	rw          sync.RWMutex                   // protects following fields.
	wheel       *TimingWheel                   // timing wheel, concurrent add task(read-lock) and advance clock only one(write-lock).
	quit        chan struct{}                  // of chan struct{}, created when first start.
	closed      bool                           // true if closed.
}

// NewTimer new timer instance. tick is 1 milliseconds, wheel size is 512.
func NewTimer(opts ...Option) *Timer {
	t := &Timer{
		tickMs:      DefaultTickMs,
		wheelSize:   DefaultWheelSize,
		wheelMask:   DefaultWheelSize - 1,
		taskCounter: atomic.Int64{},
		delayQueue:  delayqueue.NewDelayQueue(compareSpoke),
		goPool:      goroutinePool,
		quit:        closedchan,
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

// TickMs basic time tick milliseconds.
func (t *Timer) TickMs() int64 { return t.tickMs }

// WheelSize wheel size.
func (t *Timer) WheelSize() int { return t.wheelSize }

// WheelMask wheel mask.
func (t *Timer) WheelMask() int { return t.wheelMask }

// TaskCounter the total number of tasks.
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
	select {
	case <-t.quit:
		return ErrClosed
	default:
		t.rw.RLock()
		defer t.rw.RUnlock()
		t.addTask(task)
	}
	return nil
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
		quit := make(chan struct{})
		t.closed = false
		t.quit = quit
		go func() {
			for {
				spoke, exit := t.delayQueue.Take(quit)
				if exit {
					break
				}
				t.rw.Lock()
				t.wheel.advanceClock(spoke.GetExpiration())
				spoke.Flush(t.reinsert)
				t.rw.Unlock()
			}
		}()
	}
}

// Stop the timer.
func (t *Timer) Stop() {
	t.rw.Lock()
	defer t.rw.Unlock()
	if !t.closed {
		close(t.quit)
		t.closed = true
	}
}

func (t *Timer) addToDelayQueue(spoke *Spoke) {
	t.delayQueue.Add(spoke)
}

func (t *Timer) addTask(task *Task) {
	if !t.wheel.add(task) {
		if !task.Cancelled() { // already expired or cancelled
			t.goPool.Go(task.Run)
		}
	}
}

func (t *Timer) reinsert(task *Task) {
	t.addTask(task)
}
