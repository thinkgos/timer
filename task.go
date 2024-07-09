package timer

import (
	"fmt"
	"os"
	"sync"
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

// Task 是双向链表的一个元素.
type Task struct {
	delayMs   int64 // 延迟多少(初始化后不可变), 单位: ms
	job       Job   // 未来执行的工作任务
	rw        sync.RWMutex
	taskEntry *taskEntry
}

// NewTask new task with delay duration and empty job, the accuracy is milliseconds.
func NewTask(d time.Duration) *Task {
	delayMs := int64(d / time.Millisecond)
	return &Task{
		delayMs: delayMs,
		job:     emptyJob,
	}
}

// NewTaskFunc new task with delay duration and function job, the accuracy is milliseconds.
func NewTaskFunc(d time.Duration, f func()) *Task {
	return NewTask(d).WithJobFunc(f)
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

// Delay delay duration, the accuracy is milliseconds.
func (t *Task) Delay() time.Duration {
	return time.Duration(t.delayMs) * time.Millisecond
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

func (t *Task) getTaskEntry() *taskEntry {
	t.rw.RLock()
	defer t.rw.RUnlock()
	return t.taskEntry
}
