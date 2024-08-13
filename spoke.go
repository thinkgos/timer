package timer

import (
	"sync"
	"sync/atomic"
	"time"
)

// Spoke a spoke of the wheel.
type Spoke struct {
	root        taskEntry // sentinel list element, only &root, root.prev, and root.next are used
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
func (sp *Spoke) Add(te *taskEntry) {
	for done := false; !done; {
		// Remove the timer task if it is already in any other list
		// We do this outside of the sync block below to avoid deadlocking.
		// We may retry until task.list becomes null.
		te.remove()
		if te.list.Load() == nil { // fast check.
			sp.mu.Lock()
			if te.list.Load() == nil { // double check but slow.
				sp.pushBack(te)
				done = true
			}
			sp.mu.Unlock()
		}
	}
}

// Remove the specified timer task from this list
func (sp *Spoke) Remove(te *taskEntry) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if te.list.Load() == sp {
		sp.remove(te)
	}
}

func (sp *Spoke) pushBack(te *taskEntry) {
	at := sp.root.prev

	te.prev = at
	te.next = at.next
	te.prev.next = te
	te.next.prev = te
	te.list.Store(sp)
	sp.taskCounter.Add(1)
}

func (sp *Spoke) remove(te *taskEntry) {
	te.prev.next = te.next
	te.next.prev = te.prev
	te.next = nil // avoid memory leaks
	te.prev = nil // avoid memory leaks
	te.list.Store(nil)
	sp.taskCounter.Add(-1)
}

// Flush all task entries and apply the supplied function to each of them
func (sp *Spoke) Flush(f func(*taskEntry)) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	for e := sp.root.next; e != &sp.root; e = sp.root.next {
		sp.remove(e)
		f(e)
	}
	sp.SetExpiration(-1)
}

// SetExpiration the spoke's expiration time
// Returns true if the expiration time is changed
func (sp *Spoke) SetExpiration(expirationMs int64) bool {
	return sp.expiration.Swap(expirationMs) != expirationMs
}

// GetExpiration the spoke's expiration time
func (sp *Spoke) GetExpiration() int64 { return sp.expiration.Load() }

// Delay implements delayqueue.Delayed.
func (sp *Spoke) Delay() int64 {
	delay := sp.GetExpiration() - time.Now().UnixMilli()
	if delay < 0 {
		return 0
	}
	return delay
}

// CompareSpoke compare two `Spoke` with expiration.
func CompareSpoke(sp1, sp2 *Spoke) int {
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
		expirationMs: task.delay.Load()/int64(time.Millisecond) + time.Now().UnixMilli(),
	}
	task.setTaskEntry(e)
	return e
}

// ExpirationMs expiration milliseconds.
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
	return te.task.getTaskEntry() != te
}
