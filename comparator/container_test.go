package comparator

import (
	"cmp"
	"slices"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thinkgos/timer/go/heap"
)

func Test_Container_Sort(t *testing.T) {
	items := []int{5, 9, 3, 7, 10, 2, 6, 1, 8}
	wantASC := []int{1, 2, 3, 5, 6, 7, 8, 9, 10}
	wantDESC := []int{10, 9, 8, 7, 6, 5, 3, 2, 1}

	// asc
	c1 := Container[int]{
		Items:   slices.Clone(items),
		Desc:    false,
		Compare: cmp.Compare[int],
	}
	sort.Sort(c1)
	require.Equal(t, wantASC, c1.Items)

	// desc
	c2 := Container[int]{
		Items:   slices.Clone(items),
		Desc:    true,
		Compare: cmp.Compare[int],
	}
	sort.Sort(c2)
	require.Equal(t, wantDESC, c2.Items)
}

func Test_Container_Heap(t *testing.T) {
	items := []int{5, 9, 3, 7, 10, 2, 6, 1, 8}

	// min heap
	c1 := &Container[int]{
		Items:   slices.Clone(items),
		Desc:    false,
		Compare: cmp.Compare[int],
	}
	heap.Init[int](c1)
	heap.Push(c1, int(11))
	heap.Push(c1, int(12))
	for _, v := range []int{1, 2, 3, 5, 6, 7, 8, 9, 10, 11, 12} {
		require.Equal(t, v, heap.Pop(c1))
	}
	require.Zero(t, c1.Len())

	// max heap
	c2 := &Container[int]{
		Items:   slices.Clone(items),
		Desc:    true,
		Compare: cmp.Compare[int],
	}
	heap.Init(c2)
	heap.Push(c2, int(11))
	heap.Push(c2, int(12))
	for _, v := range []int{12, 11, 10, 9, 8, 7, 6, 5, 3, 2, 1} {
		require.Equal(t, v, heap.Pop(c2))
	}
	require.Zero(t, c2.Len())
}
