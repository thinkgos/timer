package comparator

import (
	"sort"

	"github.com/thinkgos/timer/go/heap"
)

var (
	_ heap.Interface[int] = (*Container[int])(nil)
	_ sort.Interface      = (*Container[int])(nil)
)

type Container[T any] struct {
	Items   []T           // container data
	Desc    bool          // asc or desc, default asc.
	Compare Comparable[T] // cmp.Compare or custom comparison
}

// Len implement heap.Interface.
func (c Container[T]) Len() int {
	return len(c.Items)
}

// Swap implement heap.Interface.
func (c Container[T]) Swap(i, j int) {
	c.Items[i], c.Items[j] = c.Items[j], c.Items[i]
}

// Less implement heap.Interface.
func (c Container[T]) Less(i, j int) bool {
	if c.Desc {
		i, j = j, i
	}
	return c.Compare(c.Items[i], c.Items[j]) < 0
}

// Push implement heap.Interface.
func (c *Container[T]) Push(x T) {
	c.Items = append(c.Items, x)
}

// Pop implement heap.Interface.
func (c *Container[T]) Pop() T {
	var zero T

	old := c.Items
	n := len(old)
	x := old[n-1]
	old[n-1] = zero // avoid memory leak
	c.Items = old[:n-1]
	return x
}
