package queue

import (
	"container/heap"

	"github.com/thinkgos/timer/comparator"
)

// PriorityQueue represents an unbounded priority queue based on a priority heap.
// It implements heap.Interface.
type PriorityQueue[T comparator.Comparable[T]] struct {
	container *comparator.Container[T]
}

// NewPriorityQueue initializes and returns an Queue, default min heap.
func NewPriorityQueue[T comparator.Comparable[T]](maxHeap bool, items ...T) *PriorityQueue[T] {
	q := &PriorityQueue[T]{
		container: &comparator.Container[T]{
			Items: items,
			Desc:  maxHeap,
		},
	}
	heap.Init(q.container)
	return q
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
		return heap.Pop(pq.container).(T), true
	}
	return val, false
}
