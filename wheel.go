package timer

type Wheel struct {
	*Timer
	tickMs int64

	interval    int64
	currentTime int64
	buckets     []*TaskList

	overflowWheel *Wheel
}

func newTimeWheel(tm *Timer, tickMs int64, startMs int64) *Wheel {
	buckets := make([]*TaskList, tm.wheelSize)
	for i := range buckets {
		buckets[i] = NewTaskList(tm.counter)
	}
	return &Wheel{
		Timer:  tm,
		tickMs: tickMs,

		interval:      tickMs * int64(tm.wheelSize),
		currentTime:   startMs - (startMs % tickMs),
		buckets:       buckets,
		overflowWheel: nil,
	}
}

func (tw *Wheel) Add(entry *TaskEntry) bool {
	if entry.hasCancelled() { // Canceled
		return false
	}
	expiration := entry.expirationMs
	if expiration < tw.currentTime+tw.tickMs { // Already expired
		return false
	}
	if expiration < tw.currentTime+tw.interval {
		// Put in its own bucket
		virtualId := expiration / tw.tickMs
		bucket := tw.buckets[int(virtualId)&(tw.wheelSize-1)]
		bucket.Add(entry)

		// Set the bucket expiration time
		if bucket.SetExpiration(virtualId * tw.tickMs) {
			// The bucket needs to be enqueued because it was an expired bucket
			// We only need to enqueue the bucket when its expiration time has changed, i.e. the wheel has advanced
			// and the previous buckets gets reused; further calls to set the expiration within the same wheel cycle
			// will pass in the same value and hence return false, thus the bucket with the same expiration will not
			// be enqueued multiple times.
			tw.delayQueue.Add(bucket)
		}
		return true
	}
	if tw.overflowWheel == nil {
		tw.overflowWheel = newTimeWheel(tw.Timer, tw.interval, tw.currentTime)
	}
	return tw.overflowWheel.Add(entry)
}

func (tw *Wheel) AdvanceClock(timeMs int64) {
	if timeMs >= tw.currentTime+tw.tickMs {
		tw.currentTime = timeMs - (timeMs % tw.tickMs)
		if tw.overflowWheel != nil {
			tw.overflowWheel.AdvanceClock(tw.currentTime)
		}
	}
}
