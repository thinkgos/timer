package timer

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/things-go/timer/delayqueue"
)

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

type Timer struct {
	tickMs      int64                          // 基本时间跨度, 单位ms
	wheelSize   int                            // 轮的大小, 2的n次方
	taskCounter *atomic.Int64                  // 任务总数
	delayQueue  *delayqueue.DelayQueue[*Spoke] // 延迟队列
	wheel       *TimingWheel                   // 时间轮
	goPool      GoPool                         // 协程池
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewTimer(opts ...Option) *Timer {
	t := &Timer{
		tickMs:      1,
		wheelSize:   32,
		taskCounter: &atomic.Int64{},
		delayQueue:  delayqueue.NewDelayQueue[*Spoke](),
		goPool:      InternalGoPool{},
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
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.wheel = newTimingWheel(t, t.tickMs, time.Now().UnixMilli())
	return t
}

func (t *Timer) TickMs() int64      { return t.tickMs }
func (t *Timer) WheelSize() int     { return t.wheelSize }
func (t *Timer) TaskCounter() int64 { return t.taskCounter.Load() }

func (t *Timer) AfterFunc(d time.Duration, f func()) *TaskEntry {
	e := NewTaskEntry(int64(d / time.Millisecond)).WithJobFunc(f)
	t.AddTask(e)
	return e
}

func (t *Timer) AddTask(e *TaskEntry) {
	if !t.wheel.Add(e) {
		if !e.isCancelled() {
			t.goPool.Go(e.Run)
		}
	}
}

func (t *Timer) reinsert(entry *TaskEntry) {
	t.AddTask(entry)
}

func (t *Timer) Start() {
	go func() {
		for {
			d := t.delayQueue.Take(t.ctx)
			if d == nil {
				break
			}
			spoke := d.(*Spoke)
			t.wheel.AdvanceClock(spoke.GetExpiration())
			spoke.Flush(t.reinsert)
		}
	}()
}

func (t *Timer) Stop() {
	if t.cancel != nil {
		t.cancel()
	}
}
