package timer

import (
	"sync"
	"sync/atomic"
	"time"
)

// Spoke a spoke of the wheel.
type Spoke struct {
	root        Task // sentinel list element, only &root, root.prev, and root.next are used
	len         int  // current list length excluding (this) sentinel element
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
func (sp *Spoke) Add(task *Task) {
	for done := false; !done; {
		// Remove the timer task if it is already in any other list
		// We do this outside of the sync block below to avoid deadlocking.
		// We may retry until task.list becomes null.
		task.removeSelf()
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
func (sp *Spoke) Remove(task *Task) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if task.list.Load() == sp {
		sp.remove(task)
	}
}

func (sp *Spoke) pushBack(task *Task) {
	at := sp.root.prev

	task.prev = at
	task.next = at.next
	task.prev.next = task
	task.next.prev = task
	task.list.Store(sp)
	sp.len++
	sp.taskCounter.Add(1)
}

func (sp *Spoke) remove(task *Task) {
	task.prev.next = task.next
	task.next.prev = task.prev
	task.next = nil // avoid memory leaks
	task.prev = nil // avoid memory leaks
	task.list.Store(nil)
	sp.len--
	sp.taskCounter.Add(-1)
}

// Front returns the first task of list l or nil if the list is empty.
func (sp *Spoke) frontTask() *Task {
	if sp.len == 0 {
		return nil
	}
	return sp.root.next
}

// Flush all task entries and apply the supplied function to each of them
func (sp *Spoke) Flush(f func(*Task)) {
	var temp *Task

	sp.mu.Lock()
	defer sp.mu.Unlock()
	for e := sp.frontTask(); e != nil; e = temp {
		temp = e.nextTask()
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

// DelayMs implements delayqueue.Delayed.
func (sp *Spoke) DelayMs() int64 {
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
