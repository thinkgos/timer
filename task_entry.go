package timer

import (
	"sync/atomic"
)

// TaskEntry consists of a schedule and the func to execute on that schedule.
type TaskEntry struct {
	// Next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).

	prev *TaskEntry
	next *TaskEntry

	// The list to which this element belongs.
	list *list

	// 到期终止ms值
	expirationMs int64
	// job is the thing that want to run.
	job Job
	// use goroutine
	useGoroutine bool
	cancelled    int32
}

// NewTaskEntry new timer
func NewTaskEntry(expirationMs int64, f func()) *TaskEntry {
	return &TaskEntry{
		expirationMs: expirationMs,
		job:          JobFunc(f),
	}
}

func (t *TaskEntry) WithGoroutine() *TaskEntry {
	t.useGoroutine = true
	return t
}

func (t *TaskEntry) Cancelled() bool {
	return atomic.LoadInt32(&t.cancelled) == 1
}

func (t *TaskEntry) Cancel() {
	atomic.StoreInt32(&t.cancelled, 1)
}

func (t *TaskEntry) Run() {
	t.job.Run()
}

// removeSelf remove self from list ,if it not on any list do nothing
func (t *TaskEntry) removeSelf() {
	if t.list != nil {
		t.list.remove(t)
	}
}
