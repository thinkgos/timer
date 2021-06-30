package timer

import (
	"log"
	"sync/atomic"
)

type Wheel struct {
	*Timer

	interval    int64
	currentTime int64
	buckets     []*TaskList

	overflowWheel *Wheel
}

func NewWheel(tm *Timer, startMs int64) *Wheel {
	buckets := make([]*TaskList, tm.wheelSize)
	for i := range buckets {
		buckets[i] = NewTaskList(tm.counter)
	}
	return &Wheel{
		Timer:       tm,
		interval:    tm.tickMs * int64(tm.wheelSize),
		currentTime: startMs - startMs%tm.tickMs,
		buckets:     buckets,
	}
}

func (tw *Wheel) Add(e *TaskEntry) bool {

	if e.Cancelled() {
		log.Println("1")
		return false
	}
	expiration := e.expirationMs
	if expiration < tw.currentTime+tw.tickMs { // Already expired
		log.Println("2")
		return false
	}
	if expiration < tw.currentTime+tw.interval {
		log.Println("3")
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
		log.Println("4")
		tw.overflowWheel = NewWheel(tw.Timer, tw.currentTime)
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
