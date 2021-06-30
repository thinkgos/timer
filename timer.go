package timer

import (
	"context"
	"time"

	"go.uber.org/atomic"

	"github.com/things-go/timer/delayqueue"
)

type Timer struct {
	tickMs     int64
	wheelSize  int
	counter    *atomic.Int64
	delayQueue *delayqueue.DelayQueue
	wheel      *Wheel
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewTimer(tickMs int64, wheelSize int) *Timer {
	tm := &Timer{
		tickMs:     tickMs,
		wheelSize:  wheelSize,
		counter:    &atomic.Int64{},
		delayQueue: delayqueue.NewDelayQueue(),
	}

	tm.wheel = NewWheel(tm)
	return tm
}

func (t *Timer) AfterFunc(d time.Duration, f func()) *TaskEntry {
	delayMs := int64(d / time.Millisecond)

	entry := NewTaskEntry(delayMs, f)

	t.addTimerTaskEntry(entry)

	return entry
}

func (t *Timer) addTimerTaskEntry(entry *TaskEntry) {
	if !t.wheel.Add(entry) {
		// Already expired or cancelled
		if !entry.Cancelled() {
			go func() {
				entry.Run()
			}()
		}
	}
}

func (t *Timer) Start() {
	go func() {
		for {
			d := t.delayQueue.Pop(context.Background())
			if d == nil {
				break
			}
			bucket := d.(*TaskList)
			t.wheel.AdvanceClock(bucket.GetExpiration())
			// bucket.Flush(t.reinsert)
		}
	}()
}

func (t *Timer) Stop() {
	if t.cancel != nil {
		t.cancel()
	}
}

func (t *Timer) Size() int64 {
	return t.counter.Load()
}
