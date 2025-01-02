package timer

import (
	"sync/atomic"
	"time"
)

// taskEntry is an element of a linked list, hold the task instance.
type taskEntry struct {
	// next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	prev         *taskEntry
	next         *taskEntry
	list         atomic.Pointer[Spoke] // The list to which this element belongs.
	expirationMs int64                 // expiration time, absolute time(immutable after first initialization), Units: ms
	task         *Task                 // the task instance.
}

func newTaskEntry(task *Task) *taskEntry {
	te := &taskEntry{
		task:         task,
		expirationMs: int64(task.Delay()/time.Millisecond) + time.Now().UnixMilli(),
	}
	task.setBelongTo(te)
	return te
}

// ExpirationMs return the expiration milliseconds.
func (te *taskEntry) ExpirationMs() int64 { return te.expirationMs }

func (te *taskEntry) remove() {
	// If remove is called when another thread is moving the entry from a task entry list to another,
	// this may fail to remove the entry due to the change of value of list. Thus, we retry until the list becomes null.
	// In a rare case, this thread sees null and exits the loop, but the other thread insert the entry to another list later.
	for currentList := te.list.Load(); currentList != nil; currentList = te.list.Load() {
		currentList.Remove(te)
	}
}

func (te *taskEntry) cancelled() bool {
	return !te.task.isBelongTo(te)
}

// activated return true if the task entry is activated.
func (te *taskEntry) activated() bool {
	return te.list.Load() != nil
}
