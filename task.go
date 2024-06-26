package timer

import (
	"fmt"
	"os"
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

// Task 是双向链表的一个元素.
type Task struct {
	// next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	prev *Task
	next *Task
	list atomic.Pointer[Spoke] // 此元素所属的列表

	// follow values associated with this element.
	delayMs      int64       // 延迟多少(初始化后不可变), 单位: ms
	expirationMs int64       // 到期时间, 绝对时间(初始化后不可变), 单位: ms
	job          Job         // 未来执行的工作任务
	hasCancelled atomic.Bool // 是否取消
}

// NewTask new task with delay duration and empty job, the accuracy is milliseconds.
func NewTask(d time.Duration) *Task {
	delayMs := int64(d / time.Millisecond)
	return &Task{
		delayMs:      delayMs,
		expirationMs: delayMs + time.Now().UnixMilli(),
		job:          emptyJob,
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
	t.hasCancelled.Store(true)
	t.removeSelf()
}

// Delay delay duration, the accuracy is milliseconds.
func (t *Task) Delay() time.Duration { return time.Duration(t.delayMs) * time.Millisecond }

// ExpirationMs expiration milliseconds.
func (t *Task) ExpirationMs() int64 { return t.expirationMs }

// Cancelled return true if the task is cancelled.
func (t *Task) Cancelled() bool { return t.hasCancelled.Load() }

// nextTask return the next task or nil.
func (t *Task) nextTask() *Task {
	if p, list := t.next, t.list.Load(); list != nil && p != &list.root {
		return p
	}
	return nil
}

func (t *Task) removeSelf() {
	// If remove is called when another thread is moving the entry from a task entry list to another,
	// this may fail to remove the entry due to the change of value of list. Thus, we retry until the list becomes null.
	// In a rare case, this thread sees null and exits the loop, but the other thread insert the entry to another list later.
	for currentList := t.list.Load(); currentList != nil; currentList = t.list.Load() {
		currentList.Remove(t)
	}
}
