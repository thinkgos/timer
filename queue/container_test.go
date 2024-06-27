package queue

import (
	"container/heap"
	"slices"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Container_Sort(t *testing.T) {
	items := []Int{5, 9, 3, 7, 10, 2, 6, 1, 8}
	wantASC := []Int{1, 2, 3, 5, 6, 7, 8, 9, 10}
	wantDESC := []Int{10, 9, 8, 7, 6, 5, 3, 2, 1}

	// asc
	c1 := Container[Int]{
		Items: slices.Clone(items),
		Desc:  false,
	}
	sort.Sort(c1)
	require.Equal(t, wantASC, c1.Items)

	// desc
	c2 := Container[Int]{
		Items: slices.Clone(items),
		Desc:  true,
	}
	sort.Sort(c2)
	require.Equal(t, wantDESC, c2.Items)
}

func Test_Container_Heap(t *testing.T) {
	items := []Int{5, 9, 3, 7, 10, 2, 6, 1, 8}

	// min heap
	c1 := &Container[Int]{
		Items: slices.Clone(items),
		Desc:  false,
	}
	heap.Init(c1)
	heap.Push(c1, Int(11))
	heap.Push(c1, Int(12))
	for _, v := range []Int{1, 2, 3, 5, 6, 7, 8, 9, 10, 11, 12} {
		require.Equal(t, v, heap.Pop(c1).(Int))
	}
	require.Zero(t, c1.Len())

	// max heap
	c2 := &Container[Int]{
		Items: slices.Clone(items),
		Desc:  true,
	}
	heap.Init(c2)
	heap.Push(c2, Int(11))
	heap.Push(c2, Int(12))
	for _, v := range []Int{12, 11, 10, 9, 8, 7, 6, 5, 3, 2, 1} {
		require.Equal(t, v, heap.Pop(c2).(Int))
	}
	require.Zero(t, c2.Len())
}
