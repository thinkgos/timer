package timer

import (
	"sync/atomic"
)

type Wheel struct {
	*Timer

	interval    int64
	currentTime int64
	buckets     []*TaskList

	overflowWheel *Wheel
}

func NewWheel(tm *Timer) *Wheel {
	buckets := make([]*TaskList, tm.wheelSize)
	for i := range buckets {
		buckets[i] = NewTaskList(tm.counter)
	}
	return &Wheel{
		Timer:   tm,
		buckets: buckets,
	}
}

func (tw *Wheel) Add(e *TaskEntry) bool {
	expiration := e.expirationMs
	if e.Cancelled() {
		// Canceled
		return false
	}
	if expiration < tw.currentTime+tw.tickMs { // Already expired
		return false
	}
	if expiration < tw.currentTime+tw.interval {
		// Put in its own bucket
		virtualId := expiration / tw.tickMs
		bucket := tw.buckets[int(virtualId)%tw.wheelSize]
		bucket.Add(e)

		// Set the bucket expiration time
		if bucket.SetExpiration(virtualId * tw.tickMs) {
			tw.delayQueue.Add(bucket)
		}
		return true
	}

	if tw.overflowWheel == nil {
		tw.overflowWheel = NewWheel(tw.Timer)
	}
	return tw.overflowWheel.Add(e)
}

func (tw *Wheel) AdvanceClock(timeMs int64) {
	currentTime := atomic.LoadInt64(&tw.currentTime)
	if timeMs >= tw.currentTime+tw.tickMs {
		currentTime = timeMs - (timeMs % tw.tickMs)
		atomic.StoreInt64(&tw.currentTime, currentTime)
		if tw.overflowWheel != nil {
			tw.overflowWheel.AdvanceClock(currentTime)
		}
	}
}
