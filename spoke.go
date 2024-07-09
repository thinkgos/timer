package timer

import (
	"sync"
	"sync/atomic"
	"time"
)

// Spoke a spoke of the wheel.
type Spoke struct {
	root        taskEntry // sentinel list element, only &root, root.prev, and root.next are used
	len         int       // current list length excluding (this) sentinel element
	taskCounter *atomic.Int64
	expiration  atomic.Int64
	mu          sync.Mutex
}

func NewSpoke(taskCounter *atomic.Int64) *Spoke {
	sp := &Spoke{
		taskCounter: taskCounter,
	}
	sp.expiration.Store(-1)
	sp.root.next = &sp.root
	sp.root.prev = &sp.root
	return sp
}

// Add the timer task to this list
func (sp *Spoke) Add(task *taskEntry) {
	for done := false; !done; {
		// Remove the timer task if it is already in any other list
		// We do this outside of the sync block below to avoid deadlocking.
		// We may retry until task.list becomes null.
		task.remove()
		if task.list.Load() == nil { // fast check.
			sp.mu.Lock()
			if task.list.Load() == nil { // double check but slow.
				sp.pushBack(task)
				done = true
			}
			sp.mu.Unlock()
		}
	}
}

// Remove the specified timer task from this list
func (sp *Spoke) Remove(task *taskEntry) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if task.list.Load() == sp {
		sp.remove(task)
	}
}

func (sp *Spoke) pushBack(task *taskEntry) {
	at := sp.root.prev

	task.prev = at
	task.next = at.next
	task.prev.next = task
	task.next.prev = task
	task.list.Store(sp)
	sp.len++
	sp.taskCounter.Add(1)
}

func (sp *Spoke) remove(task *taskEntry) {
	task.prev.next = task.next
	task.next.prev = task.prev
	task.next = nil // avoid memory leaks
	task.prev = nil // avoid memory leaks
	task.list.Store(nil)
	sp.len--
	sp.taskCounter.Add(-1)
}

// Front returns the first task of list l or nil if the list is empty.
func (sp *Spoke) frontTask() *taskEntry {
	if sp.len == 0 {
		return nil
	}
	return sp.root.next
}

// Flush all task entries and apply the supplied function to each of them
func (sp *Spoke) Flush(f func(*taskEntry)) {
	var temp *taskEntry

	sp.mu.Lock()
	defer sp.mu.Unlock()
	for e := sp.frontTask(); e != nil; e = temp {
		temp = e.nextTaskEntry()
		sp.remove(e)
		f(e)
	}
	sp.SetExpiration(-1)
}

// Set the spoke's expiration time
// Returns true if the expiration time is changed
func (sp *Spoke) SetExpiration(expirationMs int64) bool {
	return sp.expiration.Swap(expirationMs) != expirationMs
}

// Get the spoke's expiration time
func (sp *Spoke) GetExpiration() int64 { return sp.expiration.Load() }

// Delay implements delayqueue.Delayed.
func (sp *Spoke) Delay() int64 {
	delay := sp.GetExpiration() - time.Now().UnixMilli()
	if delay < 0 {
		return 0
	}
	return delay
}

// compareSpoke compare two `Spoke`.
func compareSpoke(sp1, sp2 *Spoke) int {
	v1, v2 := sp1.GetExpiration(), sp2.GetExpiration()
	if v1 < v2 {
		return -1
	}
	if v1 > v2 {
		return 1
	}
	return 0
}

// taskEntry 是双向链表的一个元素.
type taskEntry struct {
	// next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	prev *taskEntry
	next *taskEntry
	list atomic.Pointer[Spoke] // 此元素所属的列表

	expirationMs int64 // 到期时间, 绝对时间(初始化后不可变), 单位: ms
	task         *Task
}

func newTaskEntry(task *Task) *taskEntry {
	e := &taskEntry{
		task:         task,
		expirationMs: task.delayMs + time.Now().UnixMilli(),
	}
	task.setTaskEntry(e)
	return e
}

// ExpirationMs expiration milliseconds.
func (te *taskEntry) ExpirationMs() int64 { return te.expirationMs }

// nextTask return the next task or nil.
func (te *taskEntry) nextTaskEntry() *taskEntry {
	if p, list := te.next, te.list.Load(); list != nil && p != &list.root {
		return p
	}
	return nil
}

func (te *taskEntry) remove() {
	// If remove is called when another thread is moving the entry from a task entry list to another,
	// this may fail to remove the entry due to the change of value of list. Thus, we retry until the list becomes null.
	// In a rare case, this thread sees null and exits the loop, but the other thread insert the entry to another list later.
	for currentList := te.list.Load(); currentList != nil; currentList = te.list.Load() {
		currentList.Remove(te)
	}
}

func (te *taskEntry) cancelled() bool {
	return te.task.getTaskEntry() != te
}
