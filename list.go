package timer

import (
	"go.uber.org/atomic"
)

type list struct {
	root    TaskEntry
	counter *atomic.Int64 // current list length excluding (this) sentinel element
}

// newList returns an initialized list.
func newList(counter *atomic.Int64) *list {
	return new(list).Init(counter)
}

// Init initializes or clears list l.
func (l *list) Init(counter *atomic.Int64) *list {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.counter = counter
	return l
}

// Len returns the number of elements of list l.
// The complexity is O(1).
func (l *list) Len() int64 { return l.counter.Load() }

// insert inserts e after at, increments l.len, and returns e.
func (l *list) insert(e, at *TaskEntry) *TaskEntry {
	n := at.next
	at.next = e
	e.prev = at
	e.next = n
	n.prev = e
	e.list = l
	l.counter.Inc()
	return e
}

// remove removes e from its list, decrements l.len, and returns e.
func (l *list) remove(e *TaskEntry) *TaskEntry {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.list = nil
	l.counter.Dec()
	return e
}

// Front returns the first element of list l or nil if the list is empty.
func (l *list) Front() *TaskEntry {
	if l.counter.Load() == 0 {
		return nil
	}
	return l.root.next
}

// PushElementBack inserts a new element e at the back of list l and returns e.
func (l *list) PushElementBack(e *TaskEntry) *TaskEntry {
	return l.insert(e, l.root.prev)
}

// PopFront pop the first element of list l or nil if the list is empty.
func (l *list) PopFront() *TaskEntry {
	if e := l.Front(); e != nil {
		return l.remove(e)
	}
	return nil
}

// SpliceBackList inserts an other list at the back of list l.
// and then remove all the other list element
// The lists l and other may be the same. They must not be nil.
func (l *list) SpliceBackList(other *list) {
	for other.counter.Load() > 0 {
		l.PushElementBack(other.PopFront())
	}
}
