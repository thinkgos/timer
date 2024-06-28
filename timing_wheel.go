package timer

import (
	"sync"
)

type TimingWheel struct {
	timer         *Timer
	tickMs        int64        // 时间轮的基本时间跨度, 单位ms
	interval      int64        // 时间轮的总体时间跨度, tickMs * wheelSize
	currentTime   int64        // 时间轮的表盘指针, 表示当前时间轮所处的时间, 绝对时间, 单位ms.
	spokes        []*Spoke     // 时间轮的轮辐条
	overflowWheel *TimingWheel // 更高层级时间轮
	rw            sync.RWMutex
}

func newTimingWheel(t *Timer, tickMs int64, startMs int64) *TimingWheel {
	spokes := make([]*Spoke, t.wheelSize)
	for i := range spokes {
		spokes[i] = NewSpoke(&t.taskCounter)
	}
	return &TimingWheel{
		timer:       t,
		tickMs:      tickMs,
		interval:    tickMs * int64(t.wheelSize),
		currentTime: startMs - (startMs % tickMs),
		spokes:      spokes,
	}
}

// add 加到时间轮上
// true:添加成功, false: 已取消或已过期
func (tw *TimingWheel) add(task *Task) bool {
	if task.Cancelled() { // already cancelled
		return false
	}

	expiration := task.ExpirationMs()
	tw.rw.RLock()
	switch {
	case expiration < tw.currentTime+tw.tickMs: // already expired
		tw.rw.RUnlock()
		return false
	case expiration < tw.currentTime+tw.interval: // on the current time wheel
		tw.rw.RUnlock()
		// Put in its own spoke
		virtualId := expiration / tw.tickMs
		spoke := tw.spokes[int(virtualId)&tw.timer.WheelMask()]
		spoke.Add(task)

		// Set the spoke expiration time
		if spoke.SetExpiration(virtualId * tw.tickMs) {
			// The spoke needs to be enqueued because it was an expired spoke
			// We only need to enqueue the spoke when its expiration time has changed, i.e. the wheel has advanced
			// and the previous spokes gets reused; further calls to set the expiration within the same wheel cycle
			// will pass in the same value and hence return false, thus the spoke with the same expiration will not
			// be enqueued multiple times.
			tw.timer.addToDelayQueue(spoke)
		}
		return true
	default: // not on the current wheel, add a high-level time wheel.
		needInit := tw.overflowWheel == nil
		tw.rw.RUnlock()
		if needInit {
			tw.rw.Lock()
			if tw.overflowWheel == nil {
				tw.overflowWheel = newTimingWheel(tw.timer, tw.interval, tw.currentTime)
			}
			tw.rw.Unlock()
		}
		return tw.overflowWheel.add(task)
	}
}

func (tw *TimingWheel) advanceClock(timeMs int64) {
	tw.rw.Lock()
	if timeMs >= tw.currentTime+tw.tickMs {
		currentTime := timeMs - (timeMs % tw.tickMs)
		tw.currentTime = currentTime
		overflowWheel := tw.overflowWheel
		tw.rw.Unlock()
		if overflowWheel != nil {
			overflowWheel.advanceClock(currentTime)
		}
	} else {
		tw.rw.Unlock()
	}
}
