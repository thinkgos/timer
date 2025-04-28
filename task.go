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
var _ DerefTask = (*Task)(nil)
var _ Job = (*Task)(nil)

// Task timer task.
type Task struct {
	delay     atomic.Int64 // delay duration
	job       Job          // the job of future execution
	rw        sync.RWMutex // protects following fields.
	taskEntry *taskEntry   // the taskEntry to which the task belongs.
}

// NewTask new task with delay duration and an empty job, the accuracy is milliseconds.
func NewTask(d time.Duration) *Task {
	t := &Task{job: emptyJob}
	t.delay.Store(int64(d))
	return t
}

// NewTaskFunc new task with delay duration and a function job, the accuracy is milliseconds.
func NewTaskFunc(d time.Duration, f func()) *Task {
	return NewTask(d).WithJobFunc(f)
}

// NewTaskJob new task with delay duration and a job, the accuracy is milliseconds.
func NewTaskJob(d time.Duration, job Job) *Task {
	return NewTask(d).WithJob(job)
}

// WithJobFunc with a function job
func (t *Task) WithJobFunc(f func()) *Task {
	t.job = JobFunc(f)
	return t
}

// WithJob with a job
func (t *Task) WithJob(j Job) *Task {
	t.job = j
	return t
}

// DerefTask implements TaskContainer.
func (t *Task) DerefTask() *Task { return t }

// Run immediate call job. implement Job interface.
func (t *Task) Run() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintf(os.Stderr, "timer: Recovered from panic: %v\n", err)
		}
	}()
	t.job.Run()
}

// Cancel the task.
func (t *Task) Cancel() {
	t.rw.Lock()
	defer t.rw.Unlock()
	if t.taskEntry != nil {
		t.taskEntry.remove()
		t.taskEntry = nil
	}
}

// Delay return the delay duration.
func (t *Task) Delay() time.Duration {
	return time.Duration(t.delay.Load())
}

// SetDelay set a new delay duration, the accuracy is milliseconds.
// NOTE: Only effect when re-add to `Timer`, It has no effect on the task being running!
func (t *Task) SetDelay(d time.Duration) *Task {
	t.delay.Store(int64(d))
	return t
}

// Activated return true if the task is activated.
func (t *Task) Activated() bool {
	t.rw.RLock()
	defer t.rw.RUnlock()
	// why need check task entry?
	// when cancel, we will set t.taskEntry to nil,
	// but if the task is expired, only remove the task entry from the spoke.
	// so we should check the task entry..
	return t.taskEntry != nil && t.taskEntry.activated()
}

// Expiry return the milliseconds as a Unix time when the task will be expired.
// the number of milliseconds elapsed since January 1, 1970 UTC.
// the value -1 indicate the task not activated.
func (t *Task) Expiry() int64 {
	t.rw.RLock()
	defer t.rw.RUnlock()
	if t.taskEntry != nil && t.taskEntry.activated() {
		return t.taskEntry.ExpirationMs()
	}
	return -1
}

// ExpiryAt return the local time when the task will be expired.
// the zero time indicate the task not activated.
func (t *Task) ExpiryAt() time.Time {
	if ms := t.Expiry(); ms < 0 {
		return time.Time{}
	} else {
		return time.UnixMilli(ms)
	}
}

// setBelongTo set the task belongs to the task entry.
func (t *Task) setBelongTo(te *taskEntry) {
	t.rw.Lock()
	defer t.rw.Unlock()
	// if this task already belong to an existing task entry,
	// we should remove such an entry first.
	if t.taskEntry != nil && t.taskEntry != te {
		t.taskEntry.remove()
	}
	t.taskEntry = te
}

func (t *Task) isBelongTo(te *taskEntry) bool {
	t.rw.RLock()
	defer t.rw.RUnlock()
	return t.taskEntry == te
}
