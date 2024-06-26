package queue

type Comparable interface {
	// CompareTo(Comparable) should return a negative number when v1 < v2,
	// a positive number when v1 > v2 and zero when v1 == v2.
	CompareTo(Comparable) int
}
