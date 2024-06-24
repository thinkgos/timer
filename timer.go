package timer

import (
	"context"
	"sync/atomic"
	"time"

	pq "github.com/things-go/container/priorityqueue"

	"github.com/things-go/timer/delayqueue"
	"github.com/things-go/timer/goroutine"
)

type Option func(*Timer)

func WithTickMs(tickMs int64) Option {
	return func(t *Timer) {
		t.tickMs = tickMs
	}
}

func WithWheelSize(size int) Option {
	return func(t *Timer) {
		t.wheelSize = NextPowOf2(size)
	}
}

type Timer struct {
	tickMs     int64
	wheelSize  int // must pow of 2
	counter    *atomic.Int64
	delayQueue *delayqueue.DelayQueue
	wheel      *Wheel
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewTimer(opts ...Option) *Timer {
	t := &Timer{
		tickMs:     1,
		wheelSize:  256,
		counter:    &atomic.Int64{},
		delayQueue: delayqueue.NewDelayQueue(pq.WithComparator(CompareTaskList)),
		wheel:      nil,
	}
	for _, opt := range opts {
		opt(t)
	}
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.wheel = newTimeWheel(t, t.tickMs, NowMs())
	return t
}

func (t *Timer) WheelSize() int {
	return t.wheelSize
}

func (t *Timer) TickMs() int64 {
	return t.tickMs
}

func (t *Timer) Len() int64 {
	return t.counter.Load()
}

func (t *Timer) AfterFunc(d time.Duration, f func()) *TaskEntry {
	entry := NewTaskEntry(int64(d/time.Millisecond), f)
	t.addTimerTaskEntry(entry)
	return entry
}

func (t *Timer) addTimerTaskEntry(entry *TaskEntry) {
	if !t.wheel.Add(entry) {
		if !entry.hasCancelled() {
			goroutine.Go(entry.Run)
		}
	}
}

func (t *Timer) reinsert(entry *TaskEntry) {
	t.addTimerTaskEntry(entry)
}

func (t *Timer) Start() {
	go func() {
		for {
			d := t.delayQueue.Take(t.ctx)
			if d == nil {
				break
			}
			bucket := d.(*TaskList)
			t.wheel.AdvanceClock(bucket.GetExpiration())
			bucket.Flush(t.reinsert)
		}
	}()
}

func (t *Timer) Stop() {
	if t.cancel != nil {
		t.cancel()
	}
}
