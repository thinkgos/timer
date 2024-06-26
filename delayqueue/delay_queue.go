package delayqueue

import (
	"sync"
	"sync/atomic"
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
	waiting       atomic.Bool
	phantom       T
}

func NewDelayQueue[T Delayed]() *DelayQueue[T] {
	return &DelayQueue[T]{
		priorityQueue: queue.NewPriorityQueue[T](false),
		notify:        make(chan struct{}, 1),
	}
}

func (dq *DelayQueue[T]) Add(val T) {
	dq.mu.Lock()
	dq.priorityQueue.Add(val)
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

func (dq *DelayQueue[T]) Take(quit <-chan struct{}) (t T, exit bool) {
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
				dq.priorityQueue.Poll()
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
