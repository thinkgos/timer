package delayqueue

import (
	"sync"
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
	mu            sync.Mutex              // protects following fields
	priorityQueue *queue.PriorityQueue[T] // priority queue
	waiting       bool                    // mark waiting or not.
}

// NewDelayQueue new delay queue instance.
func NewDelayQueue[T Delayed](cmp comparator.Comparable[T]) *DelayQueue[T] {
	return &DelayQueue[T]{
		notify:        make(chan struct{}, 1),
		timeUnit:      time.Millisecond,
		priorityQueue: queue.NewPriorityQueueWith(cmp),
	}
}

// TimeUnit set time unit.
//
// NOTE: it is only allowed to be called at initialization, before any `Take` may
// run. `Take` reads the time unit without holding the queue's lock, so calling this
// while a `Take` runs concurrently is a data race. The timer's own queue keeps the
// default of one millisecond and never calls this method.
func (dq *DelayQueue[T]) TimeUnit(timeUnit time.Duration) *DelayQueue[T] {
	dq.timeUnit = timeUnit
	return dq
}

// Add to queue
func (dq *DelayQueue[T]) Add(val T) {
	dq.mu.Lock()
	dq.priorityQueue.Push(val)
	first, exist := dq.priorityQueue.Peek()
	wakeUp := exist && first == val && dq.waiting
	if wakeUp {
		dq.waiting = false
	}
	dq.mu.Unlock()
	if wakeUp {
		select {
		case dq.notify <- struct{}{}:
		default:
		}
	}
}

// Take from queue.
func (dq *DelayQueue[T]) Take(quit <-chan struct{}) (val T, exit bool) {
	var phantom T

	// One timer for the whole call, created on the first wait and reset on every
	// following one. A `Take` waits again each time `Add` pushes a new head, so a
	// timer per iteration is a needless allocation on a hot path.
	var t *time.Timer
	defer func() {
		if t != nil {
			t.Stop()
		}
	}()

	for {
		dq.mu.Lock()
		dq.waiting = false
		head, exist := dq.priorityQueue.Peek()
		if !exist {
			dq.waiting = true
			dq.mu.Unlock()

			select {
			case <-dq.notify:
				continue
			case <-quit:
				return phantom, true
			}
		}

		delay := head.Delay()
		if delay <= 0 {
			dq.priorityQueue.Pop()
			dq.mu.Unlock()
			return head, false
		}
		dq.waiting = true
		dq.mu.Unlock()

		d := time.Duration(delay) * dq.timeUnit
		if t == nil {
			t = time.NewTimer(d)
		} else {
			t.Reset(d)
		}
		select {
		case <-quit:
			return phantom, true
		case <-dq.notify:
		case <-t.C:
		}
		// Drain the timer so it is reusable by the next `Reset`. `Stop` reports false
		// once the timer has fired, and the value may still be sitting in the channel
		// if we woke up on `notify` instead. The drain is non-blocking on purpose: a
		// fire that has not landed in the channel yet would block a receive forever,
		// and missing it only costs one extra loop, which just re-peeks the queue.
		if !t.Stop() {
			select {
			case <-t.C:
			default:
			}
		}
	}
}

func (dq *DelayQueue[T]) Poll() (val T, exist bool) {
	var phantom T
	dq.mu.Lock()
	defer dq.mu.Unlock()
	head, exist := dq.priorityQueue.Peek()
	if exist && head.Delay() <= 0 {
		dq.priorityQueue.Pop()
		return head, true
	} else {
		return phantom, false
	}
}
