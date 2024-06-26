package timer

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/thinkgos/timer/queue"
)

type Spoke struct {
	root        Task // sentinel list element, only &root, root.prev, and root.next are used
	taskCounter *atomic.Int64
	expiration  atomic.Int64
	mu          sync.Mutex
}

func NewSpoke(taskCounter *atomic.Int64) *Spoke {
	s := &Spoke{
		taskCounter: taskCounter,
	}
	s.expiration.Store(-1)
	s.root.next = &s.root
	s.root.prev = &s.root
	return s
}

// Add a timer task to this list
func (s *Spoke) Add(task *Task) {
	for done := false; !done; {
		// Remove the timer task if it is already in any other list
		// We do this outside of the sync block below to avoid deadlocking.
		// We may retry until task.list becomes null.
		task.removeSelf()
		if task.list.Load() == nil {
			s.mu.Lock()
			at := s.root.prev

			task.prev = at
			task.next = at.next
			task.prev.next = task
			task.next.prev = task

			task.list.Store(s)
			s.taskCounter.Add(1)
			done = true
			s.mu.Unlock()
		}
	}
}

// Remove the specified timer task from this list
func (s *Spoke) Remove(task *Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.remove(task)
}

func (s *Spoke) remove(task *Task) {
	if task.list.Load() == s {
		task.prev.next = task.next
		task.next.prev = task.prev
		task.next = nil // avoid memory leaks
		task.prev = nil // avoid memory leaks
		task.list.Store(nil)
		s.taskCounter.Add(-1)
	}
}

// Flush all task entries and apply the supplied function to each of them
func (s *Spoke) Flush(f func(*Task)) {
	var temp *Task

	s.mu.Lock()
	defer s.mu.Unlock()
	for e := s.root.next; e != nil; e = temp {
		temp = e.nextTask()
		s.remove(e)
		f(e)
	}
	s.SetExpiration(-1)
}

// Set the spoke's expiration time
// Returns true if the expiration time is changed
func (s *Spoke) SetExpiration(expirationMs int64) bool {
	return s.expiration.Swap(expirationMs) != expirationMs
}

// Get the spoke's expiration time
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
