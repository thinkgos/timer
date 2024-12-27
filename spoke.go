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
