package delayqueue

import (
	"sync"
	"time"

	"github.com/things-go/timer/queue"
)

type Delayed interface {
	DelayMs() int64
	queue.Comparable
}

type DelayQueue[T Delayed] struct {
	mu            sync.Mutex
	priorityQueue *queue.PriorityQueue[T]
	notify        chan struct{}
	waiting       bool
	phantom       T
}

func NewDelayQueue[T Delayed]() *DelayQueue[T] {
	return &DelayQueue[T]{
		priorityQueue: queue.NewPriorityQueue[T](false),
		notify:        make(chan struct{}, 1),
	}
}

func (dq *DelayQueue[T]) Add(val T) {
	var wakeUp bool

	dq.mu.Lock()
	dq.priorityQueue.Add(val)
	if dq.waiting {
		first, exist := dq.priorityQueue.Peek()
		if exist && first.CompareTo(val) == 0 {
			wakeUp = true
			dq.waiting = false
		}
	}
	dq.mu.Unlock()
	if wakeUp {
		select {
		case dq.notify <- struct{}{}:
		default:
		}
	}
}

func (dq *DelayQueue[T]) Take(quit <-chan struct{}) (t T, exit bool) {
	for {
		dq.mu.Lock()
		first, exist := dq.priorityQueue.Peek()
		if !exist {
			dq.waiting = true
			dq.mu.Unlock()

			select {
			case <-dq.notify:
				continue
			case <-quit:
				return dq.phantom, true
			}
		}

		delay := first.DelayMs()
		if delay <= 0 {
			dq.priorityQueue.Poll()
			dq.mu.Unlock()
			return first, false
		}
		dq.waiting = true
		dq.mu.Unlock()
		tm := time.NewTimer(time.Duration(delay) * time.Millisecond)
		select {
		case <-quit:
			tm.Stop()
			return dq.phantom, true
		case <-dq.notify:
			tm.Stop()
		case <-tm.C:
		}
	}
}
