package timer

import (
	stdAtomic "sync/atomic"
)

// TaskEntry consists of a schedule and the func to execute on that schedule.
// TaskEntry is an element of a linked list.
type TaskEntry struct {
	// nextEntry and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	prev, next *TaskEntry
	// The list to which this element belongs.
	list *TaskList

	// follow The value stored with this element.
	delayMs      int64
	expirationMs int64
	job          Job
	useGoroutine int32
	cancelled    int32
}

// nextEntry returns the next list element or nil.
func (e *TaskEntry) nextEntry() *TaskEntry {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

func (e *TaskEntry) Remove() {
	// If remove is called when another thread is moving the entry from a task entry list to another,
	// this may fail to remove the entry due to the change of value of list. Thus, we retry until the list becomes null.
	// In a rare case, this thread sees null and exits the loop, but the other thread insert the entry to another list later.
	for currentList := e.list; currentList != nil; currentList = e.list {
		currentList.Remove(e)
	}
}

func NewTaskEntry(delayMs int64, f func()) *TaskEntry {
	return &TaskEntry{
		delayMs:      delayMs,
		expirationMs: delayMs + NowMs(),
		job:          JobFunc(f),
	}
}

func (t *TaskEntry) UseGoroutine() *TaskEntry {
	stdAtomic.StoreInt32(&t.useGoroutine, 1)
	return t
}

func (sf *TaskEntry) Run() {
	wrapRunJob(sf.job)
}

func (sf *TaskEntry) hasCancelled() bool {
	return stdAtomic.LoadInt32(&sf.cancelled) == 1
}

func (sf *TaskEntry) Cancel() {
	if sf.list != nil {
		sf.Remove()
		stdAtomic.StoreInt32(&sf.cancelled, 1)
	}
}
