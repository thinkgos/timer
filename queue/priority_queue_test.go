package queue

import (
	"cmp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PriorityQueue_Len(t *testing.T) {
	// init 3 elements
	q := NewPriorityQueue(false, 5, 6, 7, 8, 9, 10)
	require.Equal(t, 6, q.Len())
	require.False(t, q.IsEmpty())

	// Clear all elements
	q.Clear()
	require.Zero(t, q.Len())
	require.True(t, q.IsEmpty())
}

func Test_PriorityQueue_Value(t *testing.T) {
	// create priority queue
	q := NewPriorityQueueWith[int](false, cmp.Compare)
	q.Push(15)
	q.Push(19)
	q.Push(12)
	q.Push(8)
	q.Push(13)

	require.Equal(t, 5, q.Len())

	// Peek
	val, ok := q.Peek()
	require.True(t, ok)
	require.Equal(t, 8, val)
	require.Equal(t, 5, q.Len())

	// Poll
	val, ok = q.Pop()
	require.True(t, ok)
	require.Equal(t, 8, val)
	require.Equal(t, 4, q.Len())

	val, ok = q.Pop()
	require.True(t, ok)
	require.Equal(t, 12, val)
	require.Equal(t, 3, q.Len())

	// Clear
	q.Clear()

	val, ok = q.Peek()
	require.False(t, ok)
	require.Equal(t, 0, val)

	val, ok = q.Pop()
	require.False(t, ok)
	require.Equal(t, 0, val)
}

func Test_PriorityQueue_MinHeap(t *testing.T) {
	pq := NewPriorityQueue[int](false)
	pq_Test_PriorityQueue_SortImpl(t, pq, []int{15, 19, 12, 8, 13}, []int{8, 12, 13, 15, 19})
}

func Test_PriorityQueue_MaxHeap(t *testing.T) {
	pq := NewPriorityQueue[int](true)
	pq_Test_PriorityQueue_SortImpl(t, pq, []int{15, 19, 12, 8, 13}, []int{19, 15, 13, 12, 8})
}

func pq_Test_PriorityQueue_SortImpl[T comparable](t *testing.T, q *PriorityQueue[T], input, expected []T) {
	for i := 0; i < len(input); i++ {
		q.Push(input[i])
	}

	require.Equal(t, len(input), q.Len())
	for i := 0; i < len(expected); i++ {
		val, ok := q.Pop()
		assert.True(t, ok)
		assert.Equal(t, expected[i], val)
	}
	require.Zero(t, q.Len())
}

func Test_PriorityQueue_DeleteMinHeap(t *testing.T) {
	pq := NewPriorityQueue[int](false)
	pq_Test_PriorityQueue_DeleteImpl(t, pq, []int{15, 19, 12, 8, 13}, []int{8, 12, 13, 15, 19})
}

func Test_PriorityQueue_DeleteMinHeapWithComparator(t *testing.T) {
	pq := NewPriorityQueue[int](true)
	pq_Test_PriorityQueue_DeleteImpl(t, pq, []int{15, 19, 12, 8, 13}, []int{19, 15, 13, 12, 8})
}

func Test_PriorityQueue_DeleteMaxHeap(t *testing.T) {
	pq := NewPriorityQueue[int](true)
	pq_Test_PriorityQueue_DeleteImpl(t, pq, []int{15, 19, 12, 8, 13}, []int{19, 15, 13, 12, 8})
}

func pq_Test_PriorityQueue_DeleteImpl[T comparable](t *testing.T, q *PriorityQueue[T], input, expected []T) {
	for i := 0; i < len(input); i++ {
		q.Push(input[i])
	}

	for i := 0; i < len(expected); i++ {
		val, ok := q.Pop()
		assert.True(t, ok)
		assert.Equal(t, expected[i], val)
	}
	require.Zero(t, q.Len())
}
