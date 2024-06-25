package delayqueue

import (
	"context"
	"sync"
	"time"

	"github.com/things-go/timer/queue"
)

type Delayed interface {
	DelayMs() int64
	queue.Comparable
}

type DelayQueue[T Delayed] struct {
	mu      sync.Mutex
	pq      *queue.PriorityQueue[T]
	signal  chan struct{}
	waiting bool
}

func NewDelayQueue[T Delayed]() *DelayQueue[T] {
	return &DelayQueue[T]{
		pq:     queue.NewPriorityQueue[T](false),
		signal: make(chan struct{}, 1),
	}
}

func (dq *DelayQueue[T]) Add(val T) {
	var wakeup bool

	dq.mu.Lock()
	dq.pq.Add(val)
	if dq.waiting {
		first, exist := dq.pq.Peek()
		if exist && first.CompareTo(val) == 0 {
			wakeup = true
			dq.waiting = false
		}
	}
	dq.mu.Unlock()
	if wakeup {
		select {
		case dq.signal <- struct{}{}:
		default:
		}
	}
}

func (dq *DelayQueue[T]) Take(ctx context.Context) Delayed {
	for {
		dq.mu.Lock()
		first, exist := dq.pq.Peek()
		if !exist {
			dq.waiting = true
			dq.mu.Unlock()

			select {
			case <-dq.signal:
				continue
			case <-ctx.Done():
				return nil
			}
		}

		delay := first.DelayMs()
		if delay <= 0 {
			dq.pq.Poll()
			dq.mu.Unlock()
			return first
		}
		dq.waiting = true
		dq.mu.Unlock()
		tm := time.NewTimer(time.Duration(delay) * time.Millisecond)
		select {
		case <-ctx.Done():
			tm.Stop()
			return nil
		case <-dq.signal:
			tm.Stop()
		case <-tm.C:
		}
	}
}
