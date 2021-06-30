package timer

import (
	"sync"
	stdAtomic "sync/atomic"

	"go.uber.org/atomic"
)

type TaskList struct {
	root    TaskEntry     // sentinel list element, only &root, root.prev, and root.next are used
	counter *atomic.Int64 // current list length excluding (this) sentinel element

	expiration int64
	mu         sync.Mutex
}

func NewTaskList(counter *atomic.Int64) *TaskList {
	tl := &TaskList{
		counter: counter,
	}
	tl.root.next = &tl.root
	tl.root.prev = &tl.root
	return tl
}

// Add a timer task entry to this list
func (sf *TaskList) Add(e *TaskEntry) {
	for done := false; !done; {
		e.removeSelf()
		if e.list == nil {
			sf.mu.Lock()
			at := sf.root.prev

			e.prev = at
			e.next = at.next
			e.prev.next = e
			e.next.prev = e
			e.list = sf
			sf.counter.Inc()
			done = true
			sf.mu.Unlock()
		}
	}
}

// Remove the specified timer task entry from this list
func (sf *TaskList) Remove(e *TaskEntry) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.remove(e)
}

func (sf *TaskList) remove(e *TaskEntry) {
	if e.list == sf {
		e.prev.next = e.next
		e.next.prev = e.prev
		e.next = nil // avoid memory leaks
		e.prev = nil // avoid memory leaks
		e.list = nil
		sf.counter.Dec()
	}
}

// Flush all task entries and apply the supplied function to each of them
func (sf *TaskList) Flush(f func(*TaskEntry)) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	var temp *TaskEntry
	for e := sf.root.next; e != nil; e = temp {
		temp = e.nextEntry()
		sf.remove(e)
		f(e)
	}
	sf.SetExpiration(-1)
}

// Set the bucket's expiration time
// Returns true if the expiration time is changed
func (sf *TaskList) SetExpiration(expirationMs int64) bool {
	return stdAtomic.SwapInt64(&sf.expiration, expirationMs) != expirationMs
}

// Get the bucket's expiration time
func (sf *TaskList) GetExpiration() int64 {
	return stdAtomic.LoadInt64(&sf.expiration)
}

func (sf *TaskList) DelayMs() int64 {
	delay := sf.GetExpiration() - NowMs()
	if delay < 0 {
		return 0
	}
	return delay
}

func CompareTaskList(v1, v2 interface{}) int {
	d1, d2 := v1.(*TaskList).GetExpiration(), v2.(*TaskList).GetExpiration()
	if d1 < d2 {
		return -1
	}
	if d1 > d2 {
		return 1
	}
	return 0
}
