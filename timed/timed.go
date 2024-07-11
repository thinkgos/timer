package timed

import (
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/thinkgos/timer"
)

var pool timer.GoPool = wrapperAnts{}
var tim = timer.NewTimer(timer.WithGoPool(pool))

func init() {
	tim.Start()
}

// alias
type (
	Task    = timer.Task
	Job     = timer.Job
	JobFunc = timer.JobFunc
)

// TickMs Basic time tick milliseconds.
func TickMs() int64 { return tim.TickMs() }

// WheelSize wheel size.
func WheelSize() int { return tim.WheelSize() }

// TaskCounter task total number of tasks.
func TaskCounter() int64 { return tim.TaskCounter() }

// AfterFunc adds a function to the timer.
func AfterFunc(d time.Duration, f func()) (*Task, error) { return tim.AfterFunc(d, f) }

// AddTask adds a task to the timer.
func AddTask(task *Task) error { return tim.AddTask(task) }

// Started have started or not.
func Started() bool { return tim.Started() }

// Start the timer.
func Start() { tim.Start() }

// Stop the timer.
func Stop() { tim.Stop() }

// NewTask new task with delay duration and empty job, the accuracy is milliseconds.
func NewTask(d time.Duration) *Task { return timer.NewTask(d) }

// NewTaskFunc new task with delay duration and function job, the accuracy is milliseconds.
func NewTaskFunc(d time.Duration, f func()) *Task { return timer.NewTaskFunc(d, f) }

// NewTaskJob new task with delay duration and job, the accuracy is milliseconds.
func NewTaskJob(d time.Duration, j Job) *Task { return timer.NewTaskJob(d, j) }

type wrapperAnts struct{}

func (s wrapperAnts) Go(f func()) {
	Go(f)
}

// Go run a function in `ants` goroutine pool, if submit failed, fallback to use goroutine.
func Go(f func()) {
	if err := ants.Submit(f); err != nil {
		go f()
	}
}
