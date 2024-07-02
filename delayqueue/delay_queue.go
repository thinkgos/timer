package delayqueue

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/thinkgos/timer/comparator"
	"github.com/thinkgos/timer/queue"
)

type Delayed interface {
	Delay() int64
	comparable
}

type DelayQueue[T Delayed] struct {
	notify        chan struct{}           // notify channel
	phantom       T                       // phantom data for T, not any used, just placeholder for Take function, when exit.
	mu            sync.Mutex              // protects following fields
	priorityQueue *queue.PriorityQueue[T] // priority queue
	waiting       atomic.Bool             // waiting or not.
}

func NewDelayQueue[T Delayed](cmp comparator.Comparable[T]) *DelayQueue[T] {
	return &DelayQueue[T]{
		priorityQueue: queue.NewPriorityQueueWith(false, cmp),
		notify:        make(chan struct{}, 1),
	}
}

func (dq *DelayQueue[T]) Add(val T) {
	dq.mu.Lock()
	dq.priorityQueue.Push(val)
	first, exist := dq.priorityQueue.Peek()
	wakeup := exist && first == val && dq.waiting.CompareAndSwap(true, false)
	dq.mu.Unlock()
	if wakeup {
		select {
		case dq.notify <- struct{}{}:
		default:
		}
	}
}

func (dq *DelayQueue[T]) Take(timeUnit time.Duration, quit <-chan struct{}) (val T, exit bool) {
	for {
		dq.mu.Lock()
		head, exist := dq.priorityQueue.Peek()
		if !exist {
			dq.waiting.Store(true)
			dq.mu.Unlock()

			select {
			case <-dq.notify:
				continue
			case <-quit:
				return dq.phantom, true
			}
		} else {
			delay := head.Delay()
			if delay <= 0 {
				dq.priorityQueue.Pop()
				dq.mu.Unlock()
				return head, false
			}
			dq.waiting.Store(true)
			dq.mu.Unlock()
			tm := time.NewTimer(time.Duration(delay) * timeUnit)
			select {
			case <-dq.notify:
				tm.Stop()
			case <-quit:
				tm.Stop()
				return dq.phantom, true
			case <-tm.C:
				dq.waiting.Store(false)
			}
		}
	}
}
