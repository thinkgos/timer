package timer

import "sync/atomic"

type Result int

// the result of adding a task entry to the timing wheel.
const (
	Result_Success        Result = iota // success added
	Result_Canceled                     // already canceled
	Result_AlreadyExpired               // already expired
)

type TimingWheel struct {
	timer         *Timer                      // belongs to the timer.
	tickMs        int64                       // basic time span of the timing wheel, unit is milliseconds.
	interval      int64                       // the overall time span of the time wheel, tickMs * wheelSize.
	spokes        []*Spoke                    // the spokes of the timing wheel.
	currentTime   int64                       // dial pointer of timing wheel, represents the current time of the timing wheel, absolute time, unit is milliseconds.
	overflowWheel atomic.Pointer[TimingWheel] // higher-level timing wheel.
}

func newTimingWheel(t *Timer, tickMs int64, startMs int64) *TimingWheel {
	spokes := make([]*Spoke, t.wheelSize)
	for i := range spokes {
		spokes[i] = NewSpoke(&t.taskCounter)
	}
	tw := &TimingWheel{
		timer:       t,
		tickMs:      tickMs,
		interval:    tickMs * int64(t.wheelSize),
		currentTime: startMs - (startMs % tickMs),
		spokes:      spokes,
	}
	return tw
}

// add to the timing wheel.
// true: add success, false: canceled or already expired
func (tw *TimingWheel) add(te *taskEntry) Result {
	if te.cancelled() { // already cancelled
		return Result_Canceled
	}

	expiration := te.ExpirationMs()
	switch {
	case expiration < tw.currentTime+tw.tickMs: // already expired
		return Result_AlreadyExpired
	case expiration < tw.currentTime+tw.interval: // on the current time wheel
		// Put in its own spoke
		virtualId := expiration / tw.tickMs
		spoke := tw.spokes[int(virtualId)&tw.timer.WheelMask()]
		spoke.Add(te)

		// Set the spoke expiration time
		// It safe, because only change when `Timer.rw` lock. @Spoke.Add @Spoke.Flush
		// Here `Timer.rw.RLock` concurrency change safe too.
		if spoke.SetExpiration(virtualId * tw.tickMs) {
			// The spoke needs to be enqueued because it was an expired spoke
			// We only need to enqueue the spoke when its expiration time has changed, i.e. the wheel has advanced
			// and the previous spokes gets reused; further calls to set the expiration within the same wheel cycle
			// will pass in the same value and hence return false, thus the spoke with the same expiration will not
			// be enqueued multiple times.
			tw.timer.addSpokeToDelayQueue(spoke)
		}
		return Result_Success
	default: // not on the current wheel, add a high-level time wheel.
		overflowWheel := tw.overflowWheel.Load()
		if overflowWheel == nil {
			tw.overflowWheel.CompareAndSwap(nil, newTimingWheel(tw.timer, tw.interval, tw.currentTime))
			overflowWheel = tw.overflowWheel.Load()
		}
		return overflowWheel.add(te)
	}
}

func (tw *TimingWheel) advanceClock(timeMs int64) {
	if timeMs >= tw.currentTime+tw.tickMs {
		tw.currentTime = timeMs - (timeMs % tw.tickMs)
		if overflowWheel := tw.overflowWheel.Load(); overflowWheel != nil {
			overflowWheel.advanceClock(tw.currentTime)
		}
	}
}
