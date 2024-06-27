package timer

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/thinkgos/timer/delayqueue"
)

var goroutinePool = goroutine{}

type GoPool interface {
	Go(f func())
}

type goroutine struct{}

func (goroutine) Go(f func()) {
	go f()
}

type Option func(*Timer)

// WithTickMs 设置基本时间跨度
func WithTickMs(tickMs int64) Option {
	return func(t *Timer) {
		t.tickMs = tickMs
	}
}

// WithWheelSize 设置时间轮大小
func WithWheelSize(size int) Option {
	return func(t *Timer) {
		t.wheelSize = NextPowOf2(size)
	}
}

// WithGoPool 设置协程池
func WithGoPool(p GoPool) Option {
	return func(t *Timer) {
		t.goPool = p
	}
}

type Timer struct {
	tickMs      int64                          // 基本时间跨度, 单位ms
	wheelSize   int                            // 轮的大小, 2的n次方
	taskCounter atomic.Int64                   // 任务总数
	delayQueue  *delayqueue.DelayQueue[*Spoke] // 延迟队列
	wheel       *TimingWheel                   // 时间轮
	goPool      GoPool                         // 协程池
	mu          sync.Mutex                     // protects following fields
	quit        chan struct{}                  // of chan struct{}, created when first start.
	closed      bool                           // true if closed.
}

// NewTimer new timer instance. tick is 1 milliseconds, wheel size is 32.
func NewTimer(opts ...Option) *Timer {
	t := &Timer{
		tickMs:      1,
		wheelSize:   32,
		taskCounter: atomic.Int64{},
		delayQueue:  delayqueue.NewDelayQueue[*Spoke](),
		goPool:      goroutinePool,
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

func (t *Timer) TickMs() int64      { return t.tickMs }
func (t *Timer) WheelSize() int     { return t.wheelSize }
func (t *Timer) TaskCounter() int64 { return t.taskCounter.Load() }

func (t *Timer) AfterFunc(d time.Duration, f func()) *Task {
	task := NewTask(int64(d / time.Millisecond)).WithJobFunc(f)
	t.AddTask(task)
	return task
}

func (t *Timer) AddTask(task *Task) {
	if !t.wheel.Add(task) {
		if !task.cancelled() {
			t.goPool.Go(task.Run)
		}
	}
}

// Started have started or not.
func (t *Timer) Started() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return !t.closed
}

// Start the timer.
func (t *Timer) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		t.closed = false
		quit := make(chan struct{})
		t.quit = quit
		go func() {
			for {
				spoke, exit := t.delayQueue.Take(quit)
				if exit {
					break
				}
				t.wheel.AdvanceClock(spoke.GetExpiration())
				spoke.Flush(t.reinsert)
			}
		}()
	}
}

// Stop the timer.
func (t *Timer) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.closed {
		t.closed = true
		close(t.quit)
	}
}

func (t *Timer) addToDelayQueue(spoke *Spoke) {
	t.delayQueue.Add(spoke)
}

func (t *Timer) reinsert(task *Task) {
	t.AddTask(task)
}
