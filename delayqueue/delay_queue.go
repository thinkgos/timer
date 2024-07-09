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

// DelayQueue delay queue
type DelayQueue[T Delayed] struct {
	notify        chan struct{}           // notify channel
	timeUnit      time.Duration           // time unit. default 1 millisecond.
	phantom       T                       // phantom data for T, not any used, just placeholder for Take function, when exit.
	mu            sync.Mutex              // protects following fields
	priorityQueue *queue.PriorityQueue[T] // priority queue
	waiting       atomic.Bool             // waiting or not.
}

// NewDelayQueue new delay queue instance.
func NewDelayQueue[T Delayed](cmp comparator.Comparable[T]) *DelayQueue[T] {
	return &DelayQueue[T]{
		notify:        make(chan struct{}, 1),
		timeUnit:      time.Millisecond,
		priorityQueue: queue.NewPriorityQueueWith(false, cmp),
	}
}

// TimeUnit set time unit.
func (dq *DelayQueue[T]) TimeUnit(timeUnit time.Duration) *DelayQueue[T] {
	dq.timeUnit = timeUnit
	return dq
}

// Add to queue
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

// Take from queue.
func (dq *DelayQueue[T]) Take(quit <-chan struct{}) (val T, exit bool) {
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
			tm := time.NewTimer(time.Duration(delay) * dq.timeUnit)
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

func (dq *DelayQueue[T]) Poll() (val T, exist bool) {
	dq.mu.Lock()
	defer dq.mu.Unlock()
	head, exist := dq.priorityQueue.Peek()
	if exist && head.Delay() <= 0 {
		dq.priorityQueue.Pop()
		return head, true
	} else {
		return dq.phantom, false
	}
}
