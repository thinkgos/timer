package queue

type Comparable interface {
	CompareTo(Comparable) int
}
