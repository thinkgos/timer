package delayqueue

import (
	"context"
	"sync"
	"time"

	pq "github.com/things-go/container/priorityqueue"
)

type Delayed interface {
	DelayMs() int64
}

type DelayQueue struct {
	mu      sync.Mutex
	pq      *pq.Queue
	signal  chan struct{}
	waiting bool
}

func NewDelayQueue(opts ...pq.Option) *DelayQueue {
	return &DelayQueue{
		pq:     pq.New(opts...),
		signal: make(chan struct{}, 1),
	}
}

func (sf *DelayQueue) Add(d Delayed) {
	var wakeup bool

	sf.mu.Lock()
	sf.pq.Add(d)
	if sf.waiting && sf.pq.Peek().(Delayed) == d {
		wakeup = true
		sf.waiting = false
	}
	sf.mu.Unlock()
	if wakeup {
		select {
		case sf.signal <- struct{}{}:
		default:
		}
	}
}

func (sf *DelayQueue) Pop(ctx context.Context) Delayed {
	for {
		sf.mu.Lock()
		e := sf.pq.Peek()
		if e == nil {
			sf.waiting = true
			sf.mu.Unlock()

			select {
			case <-sf.signal:
				continue
			case <-ctx.Done():
				return nil
			}
		}

		first := e.(Delayed)
		delay := first.DelayMs()
		if delay <= 0 {
			sf.pq.Poll()
			sf.mu.Unlock()
			return first
		}
		sf.waiting = true
		sf.mu.Unlock()
		tm := time.NewTimer(time.Duration(delay) * time.Millisecond)
		select {
		case <-ctx.Done():
			tm.Stop()
			return nil
		case <-sf.signal:
			tm.Stop()
		case <-tm.C:
		}
	}
}
