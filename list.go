package timer

import (
	"go.uber.org/atomic"
)

// List represents a doubly linked list.
// The zero value for List is an empty list ready to use.
type list struct {
	root TaskEntry     // sentinel list element, only &root, root.prev, and root.next are used
	len  *atomic.Int64 // current list length excluding (this) sentinel element
}

// newList returns an initialized list.
func newList(counter *atomic.Int64) *list {
	return new(list).Init(counter)
}

// Init initializes or clears list l.
func (l *list) Init(counter *atomic.Int64) *list {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = counter
	return l
}

// Front returns the first element of list l or nil if the list is empty.
func (l *list) Front() *TaskEntry {
	if l.len.Load() == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *list) Back() *TaskEntry {
	if l.len.Load() == 0 {
		return nil
	}
	return l.root.prev
}

// insert inserts e after at, increments l.len, and returns e.
func (l *list) insert(e, at *TaskEntry) *TaskEntry {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	e.list = l
	l.len.Inc()
	return e
}

// remove removes e from its list, decrements l.len, and returns e.
func (l *list) remove(e *TaskEntry) *TaskEntry {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.list = nil
	l.len.Dec()
	return e
}

// Remove removes e from l if e is an element of list l.
// It returns the element value e.
// The element must not be nil.
func (l *list) Remove(e *TaskEntry) *TaskEntry {
	if e.list == l {
		// if e.list == l, l must have been initialized when e was inserted
		// in l or l == nil (e is a zero Element) and l.remove will crash
		l.remove(e)
	}
	return e
}

// PushFront inserts a new element e at the front of list l and returns e.
func (l *list) PushFront(e *TaskEntry) *TaskEntry {
	e.removeSelf() // remove self from it's list first
	return l.insert(e, &l.root)
}

// PushBack inserts a new element e at the back of list l and returns e.
func (l *list) PushBack(e *TaskEntry) *TaskEntry {
	e.removeSelf() // remove self from it's list first
	return l.insert(e, l.root.prev)
}

// PopFront pop the first element of list l or nil if the list is empty.
func (l *list) PopFront() *TaskEntry {
	if e := l.Front(); e != nil {
		return l.remove(e)
	}
	return nil
}
