package timer

import (
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

var (
	pool         GoPool = wrapperAnts{}
	defaultOnce  sync.Once
	defaultTimer *Timer
)

// DefaultTimer return the default timer, lazily initialized and started on first call.
func DefaultTimer() *Timer {
	defaultOnce.Do(func() {
		defaultTimer = NewTimer(WithGoPool(pool))
		defaultTimer.Start()
	})
	return defaultTimer
}

// TickMs return Basic time tick milliseconds.
func TickMs() int64 { return DefaultTimer().TickMs() }

// WheelSize return the wheel size.
func WheelSize() int { return DefaultTimer().WheelSize() }

// TaskCounter return the total number of tasks.
func TaskCounter() int64 { return DefaultTimer().TaskCounter() }

// AfterFunc adds a function to the timer.
func AfterFunc(d time.Duration, f func()) (*Task, error) { return DefaultTimer().AfterFunc(d, f) }

// AddTask adds a task to the timer.
func AddTask(task *Task) error { return DefaultTimer().AddTask(task) }

// AddDerefTask adds a task from DerefTask to the timer.
func AddDerefTask(task DerefTask) error { return DefaultTimer().AddDerefTask(task) }

// Started have started or not.
func Started() bool { return DefaultTimer().Started() }

// Start the timer.
func Start() { DefaultTimer().Start() }

// Stop the timer.
func Stop() { DefaultTimer().Stop() }

type wrapperAnts struct{}

func (wrapperAnts) Go(f func()) {
	Go(f)
}

// Go run a function in `ants` goroutine pool, if submit failed, fallback to use goroutine.
func Go(f func()) {
	if err := ants.Submit(f); err != nil {
		go f()
	}
}
