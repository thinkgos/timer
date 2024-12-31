// `timed` is a global `timer` instance, that tick is 1ms. wheel size is 512,
// use [ants](https://github.com/panjf2000/ants) goroutine pool.
//
// Deprecated: As of v0.7.0, the same functionality is now provided
// by timer, and those implementations
package timed

import (
	"time"

	"github.com/thinkgos/timer"
)

// alias
type (
	Task    = timer.Task
	Job     = timer.Job
	JobFunc = timer.JobFunc
)

// Timer return the default timer.
//
// Deprecated: As of v0.7.0, use timer.DefaultTimer instead.
func Timer() *timer.Timer { return timer.DefaultTimer() }

// TickMs return Basic time tick milliseconds.
//
// Deprecated: As of v0.7.0, use timer.TickMs instead.
func TickMs() int64 { return timer.TickMs() }

// WheelSize return the wheel size.
//
// Deprecated: As of v0.7.0, use timer.WheelSize instead.
func WheelSize() int { return timer.WheelSize() }

// TaskCounter return the total number of tasks.
//
// Deprecated: As of v0.7.0, use timer.TaskCounter instead.
func TaskCounter() int64 { return timer.TaskCounter() }

// AfterFunc adds a function to the timer.
//
// Deprecated: As of v0.7.0, use timer.AfterFunc instead.
func AfterFunc(d time.Duration, f func()) (*Task, error) { return timer.AfterFunc(d, f) }

// AddTask adds a task to the timer.
//
// Deprecated: As of v0.7.0, use timer.AddTask instead.
func AddTask(task *Task) error { return timer.AddTask(task) }

// Started have started or not.
//
// Deprecated: As of v0.7.0, use timer.Started instead.
func Started() bool { return timer.Started() }

// Start the timer.
//
// Deprecated: As of v0.7.0, use timer.Start instead.
func Start() { timer.Start() }

// Stop the timer.
//
// Deprecated: As of v0.7.0, use timer.Stop instead.
func Stop() { timer.Stop() }

// NewTask new task with delay duration and an empty job, the accuracy is milliseconds.
//
// Deprecated: As of v0.7.0, use timer.NewTask instead.
func NewTask(d time.Duration) *Task { return timer.NewTask(d) }

// NewTaskFunc new task with delay duration and a function job, the accuracy is milliseconds.
//
// Deprecated: As of v0.7.0, use timer.NewTaskFunc instead.
func NewTaskFunc(d time.Duration, f func()) *Task { return timer.NewTaskFunc(d, f) }

// NewTaskJob new task with delay duration and a job, the accuracy is milliseconds.
//
// Deprecated: As of v0.7.0, use timer.NewTaskJob instead.
func NewTaskJob(d time.Duration, j Job) *Task { return timer.NewTaskJob(d, j) }
