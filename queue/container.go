package queue

import (
	"container/heap"
	"sort"
)

var _ heap.Interface = (*Container[Int])(nil)
var _ sort.Interface = (*Container[Int])(nil)

type Container[T Comparable] struct {
	Items []T  // container data
	Desc  bool // asc or desc, default asc.
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
	return c.Items[i].CompareTo(c.Items[j]) < 0
}

// Push implement heap.Interface.
func (c *Container[T]) Push(x any) {
	c.Items = append(c.Items, x.(T))
}

// Pop implement heap.Interface.
func (c *Container[T]) Pop() any {
	var zero T

	old := c.Items
	n := len(old)
	x := old[n-1]
	old[n-1] = zero // avoid memory leak
	c.Items = old[:n-1]
	return x
}
