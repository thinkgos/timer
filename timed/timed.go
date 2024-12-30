package timed

import (
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/thinkgos/timer"
)

var pool timer.GoPool = wrapperAnts{}
var defaultTimer = timer.NewTimer(timer.WithGoPool(pool))

func init() {
	defaultTimer.Start()
}

// alias
type (
	Task    = timer.Task
	Job     = timer.Job
	JobFunc = timer.JobFunc
)

// Timer return the default timer.
func Timer() *timer.Timer { return defaultTimer }

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

// Started have started or not.
func Started() bool { return defaultTimer.Started() }

// Start the timer.
func Start() { defaultTimer.Start() }

// Stop the timer.
func Stop() { defaultTimer.Stop() }

// NewTask new task with delay duration and an empty job, the accuracy is milliseconds.
func NewTask(d time.Duration) *Task { return timer.NewTask(d) }

// NewTaskFunc new task with delay duration and a function job, the accuracy is milliseconds.
func NewTaskFunc(d time.Duration, f func()) *Task { return timer.NewTaskFunc(d, f) }

// NewTaskJob new task with delay duration and a job, the accuracy is milliseconds.
func NewTaskJob(d time.Duration, j Job) *Task { return timer.NewTaskJob(d, j) }

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
