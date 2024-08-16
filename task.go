package timer

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Job job interface
type Job interface {
	Run()
}

// JobFunc job function
type JobFunc func()

// Run implement job interface
func (f JobFunc) Run() { f() }

var emptyJob = JobFunc(func() {})

// Task timer task.
type Task struct {
	delay     atomic.Int64 // delay duration
	job       Job          // the job of future execution
	rw        sync.RWMutex
	taskEntry *taskEntry
}

// NewTask new task with delay duration and empty job, the accuracy is milliseconds.
func NewTask(d time.Duration) *Task {
	t := &Task{job: emptyJob}
	t.delay.Store(int64(d))
	return t
}

// NewTaskFunc new task with delay duration and function job, the accuracy is milliseconds.
func NewTaskFunc(d time.Duration, f func()) *Task {
	return NewTask(d).WithJobFunc(f)
}

// NewTaskJob new task with delay duration and job, the accuracy is milliseconds.
func NewTaskJob(d time.Duration, job Job) *Task {
	return NewTask(d).WithJob(job)
}

// WithJobFunc with function job
func (t *Task) WithJobFunc(f func()) *Task {
	t.job = JobFunc(f)
	return t
}

// WithJob with job
func (t *Task) WithJob(j Job) *Task {
	t.job = j
	return t
}

// Run immediate call job.
func (t *Task) Run() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintf(os.Stderr, "timer: Recovered from panic: %v\n", err)
		}
	}()
	t.job.Run()
}

// Cancel the task
func (t *Task) Cancel() {
	t.rw.Lock()
	defer t.rw.Unlock()
	if t.taskEntry != nil {
		t.taskEntry.remove()
		t.taskEntry = nil
	}
}

// Delay delay duration.
func (t *Task) Delay() time.Duration {
	return time.Duration(t.delay.Load())
}

// SetDelay set delay duration, the accuracy is milliseconds.
// NOTE: only effect when re-add to `Timer`, It has no effect on the task being running!
func (t *Task) SetDelay(d time.Duration) *Task {
	t.delay.Store(int64(d))
	return t
}

// Activated return true if the task is activated.
func (t *Task) Activated() bool {
	t.rw.RLock()
	defer t.rw.RUnlock()
	return t.taskEntry != nil
}

func (t *Task) setTaskEntry(entry *taskEntry) {
	t.rw.Lock()
	defer t.rw.Unlock()
	// if this task is already held by an existing task entry,
	// we will remove such an entry first.
	if t.taskEntry != nil && t.taskEntry != entry {
		t.taskEntry.remove()
	}
	t.taskEntry = entry
}

func (t *Task) equalToTaskEntry(te *taskEntry) bool {
	t.rw.RLock()
	defer t.rw.RUnlock()
	return t.taskEntry == te
}
