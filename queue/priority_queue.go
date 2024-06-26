package queue

import (
	"container/heap"
)

// PriorityQueue represents an unbounded priority queue based on a priority heap.
// It implements heap.Interface.
type PriorityQueue[T Comparable] struct {
	data *Container[T]
}

// NewPriorityQueue initializes and returns an Queue, default min heap.
func NewPriorityQueue[T Comparable](maxHeap bool, items ...T) *PriorityQueue[T] {
	q := &PriorityQueue[T]{
		data: &Container[T]{
			Items:   items,
			Reverse: maxHeap,
		},
	}
	heap.Init(q.data)
	return q
}

// Len returns the length of this priority queue.
func (pq *PriorityQueue[T]) Len() int { return pq.data.Len() }

// IsEmpty returns true if this list contains no elements.
func (pq *PriorityQueue[T]) IsEmpty() bool { return pq.Len() == 0 }

// Clear removes all the elements from this priority queue.
func (pq *PriorityQueue[T]) Clear() { pq.data.Items = make([]T, 0) }

// Add inserts the specified element into this priority queue.
func (pq *PriorityQueue[T]) Add(item T) {
	heap.Push(pq.data, item)
}

// Peek retrieves, but does not remove, the head of this queue, or return nil if this queue is empty.
func (pq *PriorityQueue[T]) Peek() (val T, exist bool) {
	if pq.Len() > 0 {
		return pq.data.Items[0], true
	}
	return val, false
}

// Poll retrieves and removes the head of the this queue, or return nil if this queue is empty.
func (pq *PriorityQueue[T]) Poll() (val T, exist bool) {
	if pq.Len() > 0 {
		return heap.Pop(pq.data).(T), true
	}
	return val, false
}

// Contains returns true if this queue contains the specified element.
func (pq *PriorityQueue[T]) Contains(val T) bool { return pq.indexOf(val) >= 0 }

// Remove a single instance of the specified element from this queue, if it is present.
// It returns false if the target value isn't present, otherwise returns true.
func (pq *PriorityQueue[T]) Remove(val T) {
	if idx := pq.indexOf(val); idx >= 0 {
		heap.Remove(pq.data, idx)
	}
}

func (pq *PriorityQueue[T]) indexOf(val T) int {
	for i := 0; i < pq.Len(); i++ {
		if val.CompareTo(pq.data.Items[i]) == 0 {
			return i
		}
	}
	return -1
}
