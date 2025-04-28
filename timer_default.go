package timer

import (
	"time"

	"github.com/panjf2000/ants/v2"
)

var pool GoPool = wrapperAnts{}
var defaultTimer = NewTimer(WithGoPool(pool))

func init() {
	defaultTimer.Start()
}

// Timer return the default timer.
func DefaultTimer() *Timer { return defaultTimer }

// TickMs return Basic time tick milliseconds.
func TickMs() int64 { return defaultTimer.TickMs() }

// WheelSize return the wheel size.
func WheelSize() int { return defaultTimer.WheelSize() }

// TaskCounter return the total number of tasks.
func TaskCounter() int64 { return defaultTimer.TaskCounter() }

// AfterFunc adds a function to the timer.
func AfterFunc(d time.Duration, f func()) (*Task, error) { return defaultTimer.AfterFunc(d, f) }

// AddTask adds a task to the timer.
func AddTask(task *Task) error { return defaultTimer.AddTask(task) }

// AddDerefTask adds a task from TaskContainer to the timer.
func AddDerefTask(task DerefTask) error { return defaultTimer.AddDerefTask(task) }

// Started have started or not.
func Started() bool { return defaultTimer.Started() }

// Start the timer.
func Start() { defaultTimer.Start() }

// Stop the timer.
func Stop() { defaultTimer.Stop() }

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
