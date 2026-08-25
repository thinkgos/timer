package timer

import (
	"sync"
	"sync/atomic"
	"time"
)

var _ DerefTask = (*Task)(nil)
var _ Job = (*Task)(nil)

// Task timer task.
type Task struct {
	delay     atomic.Int64 // delay duration.
	job       Schedule     // the job of future execution, assign only at initialization. @Task.WithScheduleJob
	rw        sync.RWMutex // protects following fields.
	taskEntry *taskEntry   // the taskEntry to which the task belongs.
}

// NewScheduleTask new task with delay duration and a schedule job.
func NewScheduleTask(d time.Duration, sj Schedule) *Task {
	t := &Task{job: sj}
	t.delay.Store(int64(d))
	return t
}

// NewOneshotTask new task with delay duration and a job, the job will run once.
// see [timer.Oneshot]
func NewOneshotTask(d time.Duration, job Job) *Task {
	return NewScheduleTask(d, Oneshot(job))
}

// NewOneshotTaskFunc new task with delay duration and a function, the function will run once.
func NewOneshotTaskFunc(d time.Duration, f func()) *Task {
	return NewOneshotTask(d, JobFunc(f))
}

// NewPeriodicTask new task with delay duration and a job, the job will run periodically.
// see [timer.Periodic]
func NewPeriodicTask(d time.Duration, job Job) *Task {
	j := Periodic(d, job)
	return NewScheduleTask(j.NextDelay(), j)
}

// NewPeriodicTaskFunc new task with delay duration and a function, the job will run periodically.
func NewPeriodicTaskFunc(d time.Duration, f func()) *Task {
	return NewPeriodicTask(d, JobFunc(f))
}

// NewCrontabTask new task with crontab spec and a job, the job will run periodically.
func NewCrontabTask(spec string, job Job) (*Task, error) {
	j, err := Crontab(spec, job)
	if err != nil {
		return nil, err
	}
	return NewScheduleTask(j.NextDelay(), j), nil
}

// NewCrontabTaskFunc new task with crontab spec and a function, the job will run periodically.
func NewCrontabTaskFunc(spec string, f func()) (*Task, error) {
	return NewCrontabTask(spec, JobFunc(f))
}

// NewTask new task with delay duration and an empty job, the accuracy is milliseconds.
func NewTask(d time.Duration) *Task {
	return NewScheduleTask(d, emptyJob)
}

// NewTaskJob new task with delay duration and a job, the accuracy is milliseconds.
//
// Deprecated: this value is simply [timer.NewOneshotTask].
func NewTaskJob(d time.Duration, job Job) *Task {
	return NewOneshotTask(d, job)
}

// NewTaskFunc new task with delay duration and a function job, the accuracy is milliseconds.
//
// Deprecated: this value is simply [timer.NewOneshotTaskFunc].
func NewTaskFunc(d time.Duration, f func()) *Task {
	return NewOneshotTaskFunc(d, f)
}

// WithJobFunc with a function, the function will be wrapped in an OneshotJob.
//
// Deprecated: this value is simply [timer.WithScheduleJob(NewOneshotJob(JobFunc(f)))].
func (t *Task) WithJobFunc(f func()) *Task {
	return t.WithScheduleJob(Oneshot(JobFunc(f)))
}

// WithJob with a job, the job will be wrapped in an OneshotJob.
//
// Deprecated: this value is simply [timer.WithScheduleJob(NewOneshotJob(j))].
func (t *Task) WithJob(j Job) *Task {
	return t.WithScheduleJob(Oneshot(j))
}

// WithScheduleJob with a schedule job.
//
// NOTE: it is NOT safe for concurrent use, assign the job only at initialization,
// before the task is added to a `Timer`. Once added, the job is read from the
// goroutine pool (@Task.Run, @Timer.addTaskEntry), so assigning it afterwards is a
// data race. `Task.job` therefore needs no lock. To change the job, use a new task.
func (t *Task) WithScheduleJob(j Schedule) *Task {
	t.job = j
	return t
}

// DerefTask implements TaskContainer.
func (t *Task) DerefTask() *Task { return t }

// Run immediate call job. implement Job interface.
func (t *Task) Run() {
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
// It must be greater than 0, otherwise `Timer.AddTask` reports `ErrInvalidDelay`.
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
