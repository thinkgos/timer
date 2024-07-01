package delayqueue

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/thinkgos/timer/comparator"
	"github.com/thinkgos/timer/queue"
)

type Delayed[T any] interface {
	DelayMs() int64
	comparator.Comparable[T]
}

type DelayQueue[T Delayed[T]] struct {
	notify        chan struct{}           // notify channel
	phantom       T                       // phantom data for T, not any used, just placeholder for Take function, when exit.
	mu            sync.Mutex              // protects following fields
	priorityQueue *queue.PriorityQueue[T] // priority queue
	waiting       atomic.Bool             // waiting or not.
}

func NewDelayQueue[T Delayed[T]]() *DelayQueue[T] {
	return &DelayQueue[T]{
		priorityQueue: queue.NewPriorityQueue[T](false),
		notify:        make(chan struct{}, 1),
	}
}

func (dq *DelayQueue[T]) Add(val T) {
	dq.mu.Lock()
	dq.priorityQueue.Push(val)
	first, exist := dq.priorityQueue.Peek()
	wakeup := exist && first.CompareTo(val) == 0 && dq.waiting.CompareAndSwap(true, false)
	dq.mu.Unlock()
	if wakeup {
		select {
		case dq.notify <- struct{}{}:
		default:
		}
	}
}

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
			delay := head.DelayMs()
			if delay <= 0 {
				dq.priorityQueue.Pop()
				dq.mu.Unlock()
				return head, false
			}
			dq.waiting.Store(true)
			dq.mu.Unlock()
			tm := time.NewTimer(time.Duration(delay) * time.Millisecond)
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
