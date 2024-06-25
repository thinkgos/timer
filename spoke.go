package timer

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/things-go/timer/queue"
)

type Spoke struct {
	root        TaskEntry     // sentinel list element, only &root, root.prev, and root.next are used
	taskCounter *atomic.Int64 // current list length excluding (this) sentinel element

	expiration atomic.Int64
	mu         sync.Mutex
}

func NewSpoke(counter *atomic.Int64) *Spoke {
	s := &Spoke{
		taskCounter: counter,
	}
	s.root.next = &s.root
	s.root.prev = &s.root
	return s
}

// Add a timer task entry to this list
func (s *Spoke) Add(e *TaskEntry) {
	for done := false; !done; {
		e.removeSelf()
		if e.list == nil {
			s.mu.Lock()
			at := s.root.prev

			e.prev = at
			e.next = at.next
			e.prev.next = e
			e.next.prev = e
			e.list = s
			s.taskCounter.Add(1)
			done = true
			s.mu.Unlock()
		}
	}
}

// Remove the specified timer task entry from this list
func (s *Spoke) Remove(e *TaskEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.remove(e)
}

func (s *Spoke) remove(e *TaskEntry) {
	if e.list == s {
		e.prev.next = e.next
		e.next.prev = e.prev
		e.next = nil // avoid memory leaks
		e.prev = nil // avoid memory leaks
		e.list = nil
		s.taskCounter.Add(-1)
	}
}

// Flush all task entries and apply the supplied function to each of them
func (s *Spoke) Flush(f func(*TaskEntry)) {
	var temp *TaskEntry

	s.mu.Lock()
	defer s.mu.Unlock()
	for e := s.root.next; e != nil; e = temp {
		temp = e.nextEntry()
		s.remove(e)
		f(e)
	}
	s.SetExpiration(-1)
}

// Set the bucket's expiration time
// Returns true if the expiration time is changed
func (s *Spoke) SetExpiration(expirationMs int64) bool {
	return s.expiration.Swap(expirationMs) != expirationMs
}

// Get the bucket's expiration time
func (s *Spoke) GetExpiration() int64 { return s.expiration.Load() }

func (s *Spoke) DelayMs() int64 {
	delay := s.GetExpiration() - time.Now().UnixMilli()
	if delay < 0 {
		return 0
	}
	return delay
}

func (s *Spoke) CompareTo(s2 queue.Comparable) int {
	d, d2 := s.GetExpiration(), s2.(*Spoke).GetExpiration()
	if d < d2 {
		return -1
	}
	if d > d2 {
		return 1
	}
	return 0
}
