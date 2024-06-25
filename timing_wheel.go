package timer

import "sync/atomic"

type TimingWheel struct {
	timer         *Timer
	tickMs        int64                       // 时间轮的基本时间跨度, 单位ms
	interval      int64                       // 时间轮的总体时间跨度, tickMs * wheelSize
	currentTime   int64                       // 时间轮的表盘指针, 表示当前时间轮所处的时间, 绝对时间, 单位ms.
	spokes        []*Spoke                    // 时间轮的轮辐条
	overflowWheel atomic.Pointer[TimingWheel] // 更高层级时间轮
}

func newTimingWheel(t *Timer, tickMs int64, startMs int64) *TimingWheel {
	spokes := make([]*Spoke, t.wheelSize)
	for i := range spokes {
		spokes[i] = NewSpoke(t.taskCounter)
	}
	return &TimingWheel{
		timer:       t,
		tickMs:      tickMs,
		interval:    tickMs * int64(t.wheelSize),
		currentTime: startMs - (startMs % tickMs),
		spokes:      spokes,
	}
}

// Add 加到时间轮上
// true:添加成功, false: 已取消或已过期
func (tw *TimingWheel) Add(task *Task) bool {
	if task.cancelled() { // 已取消
		return false
	}
	expiration := task.expirationMs
	if expiration < tw.currentTime+tw.tickMs { // 已经过期了
		return false
	}
	if expiration < tw.currentTime+tw.interval { // 在当前时间轮上
		// Put in its own spoke
		virtualId := expiration / tw.tickMs
		spoke := tw.spokes[int(virtualId)&(tw.timer.WheelSize()-1)]
		spoke.Add(task)

		// Set the bucket expiration time
		if spoke.SetExpiration(virtualId * tw.tickMs) {
			// The bucket needs to be enqueued because it was an expired bucket
			// We only need to enqueue the bucket when its expiration time has changed, i.e. the wheel has advanced
			// and the previous buckets gets reused; further calls to set the expiration within the same wheel cycle
			// will pass in the same value and hence return false, thus the bucket with the same expiration will not
			// be enqueued multiple times.
			tw.timer.addToDelayQueue(spoke)
		}
		return true
	}
	// 不在当前轮上, 加入高一级时间轮.
	overflowWheel := tw.overflowWheel.Load()
	if overflowWheel == nil {
		tw.overflowWheel.CompareAndSwap(nil, newTimingWheel(tw.timer, tw.interval, tw.currentTime))
		overflowWheel = tw.overflowWheel.Load()
	}
	return overflowWheel.Add(task)
}

func (tw *TimingWheel) AdvanceClock(timeMs int64) {
	if timeMs >= tw.currentTime+tw.tickMs {
		tw.currentTime = timeMs - (timeMs % tw.tickMs)
		overflowWheel := tw.overflowWheel.Load()
		if overflowWheel != nil {
			overflowWheel.AdvanceClock(tw.currentTime)
		}
	}
}
