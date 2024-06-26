package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PriorityQueue_Len(t *testing.T) {
	// init 3 elements
	q := NewPriorityQueue[Int](false, 5, 6, 7, 8, 9, 10)
	require.Equal(t, 6, q.Len())
	require.False(t, q.IsEmpty())
	// remove one element
	t.Log(q.indexOf(7))
	q.Remove(7)
	require.Equal(t, 5, q.Len())

	// remove one element not exist
	q.Remove(10000)
	require.Equal(t, 5, q.Len())

	// Clear all elements
	q.Clear()
	require.Zero(t, q.Len())
	require.True(t, q.IsEmpty())

	// remove one element if not any element in queue
	q.Remove(10000)
}

func Test_PriorityQueue_Value(t *testing.T) {
	// create priority queue
	q := NewPriorityQueue[Int](false)
	q.Add(15)
	q.Add(19)
	q.Add(12)
	q.Add(8)
	q.Add(13)

	require.Equal(t, 5, q.Len())

	// Peek
	val, ok := q.Peek()
	require.True(t, ok)
	require.Equal(t, 8, int(val))
	require.Equal(t, 5, q.Len())

	// Contains
	require.True(t, q.Contains(12))
	require.False(t, q.Contains(10000))

	// Poll
	val, ok = q.Poll()
	require.True(t, ok)
	require.Equal(t, 8, int(val))
	require.Equal(t, 4, q.Len())

	val, ok = q.Poll()
	require.True(t, ok)
	require.Equal(t, 12, int(val))
	require.Equal(t, 3, q.Len())

	// Contains (again)
	require.False(t, q.Contains(12))
	require.False(t, q.Contains(10000))

	// Remove
	require.True(t, q.Contains(15))
	q.Remove(15)
	require.False(t, q.Contains(15))

	// Clear
	q.Clear()

	val, ok = q.Peek()
	require.False(t, ok)
	require.Equal(t, 0, int(val))

	val, ok = q.Poll()
	require.False(t, ok)
	require.Equal(t, 0, int(val))
}

func Test_PriorityQueue_MinHeap(t *testing.T) {
	pq := NewPriorityQueue[Int](false)
	pq_Test_PriorityQueue_SortImpl(t, pq, []Int{15, 19, 12, 8, 13}, []Int{8, 12, 13, 15, 19})
}

func Test_PriorityQueue_MaxHeap(t *testing.T) {
	pq := NewPriorityQueue[Int](true)
	pq_Test_PriorityQueue_SortImpl(t, pq, []Int{15, 19, 12, 8, 13}, []Int{19, 15, 13, 12, 8})
}

func pq_Test_PriorityQueue_SortImpl[T Comparable](t *testing.T, q *PriorityQueue[T], input, expected []T) {
	for i := 0; i < len(input); i++ {
		q.Add(input[i])
	}

	require.Equal(t, len(input), q.Len())
	for i := 0; i < len(expected); i++ {
		val, ok := q.Poll()
		assert.True(t, ok)
		assert.Equal(t, expected[i], val)
	}
	require.Zero(t, q.Len())
}

func Test_PriorityQueue_DeleteMinHeap(t *testing.T) {
	pq := NewPriorityQueue[Int](false)
	pq_Test_PriorityQueue_DeleteImpl(t, pq, []Int{15, 19, 12, 8, 13}, []Int{8, 12, 13, 15}, 19)
}

func Test_PriorityQueue_DeleteMinHeapWithComparator(t *testing.T) {
	pq := NewPriorityQueue[Int](true)
	pq_Test_PriorityQueue_DeleteImpl(t, pq, []Int{15, 19, 12, 8, 13}, []Int{19, 13, 12, 8}, 15)
}

func Test_PriorityQueue_DeleteMaxHeap(t *testing.T) {
	pq := NewPriorityQueue[Int](true)
	pq_Test_PriorityQueue_DeleteImpl(t, pq, []Int{15, 19, 12, 8, 13}, []Int{19, 15, 13, 8}, 12)
}

func pq_Test_PriorityQueue_DeleteImpl[T Comparable](t *testing.T, q *PriorityQueue[T], input, expected []T, val T) {
	for i := 0; i < len(input); i++ {
		q.Add(input[i])
	}

	q.Remove(val)
	require.Equal(t, len(input)-1, q.Len())
	assert.False(t, q.Contains(val))
	for i := 0; i < len(expected); i++ {
		val, ok := q.Poll()
		assert.True(t, ok)
		assert.Equal(t, expected[i], val)
	}
	require.Zero(t, q.Len())
}
