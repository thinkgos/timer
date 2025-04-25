package queue

import (
	"cmp"

	"github.com/thinkgos/timer/comparator"
	"github.com/thinkgos/timer/go/heap"
)

// PriorityQueue represents an unbounded priority queue based on a priority heap.
// It implements heap.Interface.
type PriorityQueue[T comparable] struct {
	container *comparator.Container[T]
}

// Option customize the PriorityQueue
type Option[T comparable] func(*PriorityQueue[T])

// WithMaxHeap customize max heap
func WithMaxHeap[T comparable]() Option[T] {
	return func(pq *PriorityQueue[T]) {
		pq.container.Desc = true
	}
}

// WithCompare customize comparator.Comparable
func WithCompare[T comparable](cmp comparator.Comparable[T]) Option[T] {
	return func(pq *PriorityQueue[T]) {
		pq.container.Compare = cmp
	}
}

// WithItems customize initialize items
func WithItems[T comparable](items ...T) Option[T] {
	return func(pq *PriorityQueue[T]) {
		pq.container.Items = items
	}
}

// NewPriorityQueue initializes and returns a priority Queue, default min heap.
func NewPriorityQueue[T cmp.Ordered](opts ...Option[T]) *PriorityQueue[T] {
	pq := &PriorityQueue[T]{
		container: &comparator.Container[T]{
			Items:   []T{},
			Desc:    false,
			Compare: cmp.Compare[T],
		},
	}
	for _, f := range opts {
		f(pq)
	}
	heap.Init(pq.container)
	return pq
}

// NewPriorityQueueWith initializes and returns a priority Queue, default min heap.
func NewPriorityQueueWith[T comparable](cmp comparator.Comparable[T], opts ...Option[T]) *PriorityQueue[T] {
	pq := &PriorityQueue[T]{
		container: &comparator.Container[T]{
			Items:   []T{},
			Desc:    false,
			Compare: cmp,
		},
	}
	for _, f := range opts {
		f(pq)
	}
	heap.Init(pq.container)
	return pq
}

// Len returns the length of this priority queue.
func (pq *PriorityQueue[T]) Len() int { return pq.container.Len() }

// IsEmpty returns true if this list contains no elements.
func (pq *PriorityQueue[T]) IsEmpty() bool { return pq.Len() == 0 }

// Clear removes all the elements from this priority queue.
func (pq *PriorityQueue[T]) Clear() { pq.container.Items = make([]T, 0) }

// Push inserts the specified element into this priority queue.
func (pq *PriorityQueue[T]) Push(item T) { heap.Push(pq.container, item) }

// Peek retrieves, but does not remove, the head of this queue, or return nil if this queue is empty.
func (pq *PriorityQueue[T]) Peek() (val T, exist bool) {
	if pq.Len() > 0 {
		return pq.container.Items[0], true
	}
	return val, false
}

// Pop retrieves and removes the head of the this queue, or return nil if this queue is empty.
func (pq *PriorityQueue[T]) Pop() (val T, exist bool) {
	if pq.Len() > 0 {
		return heap.Pop(pq.container), true
	}
	return val, false
}
