package queue

// Comparable method CompareTo(Comparable) should return a negative number when v1 < v2,
// a positive number when v1 > v2 and zero when v1 == v2.
type Comparable interface {
	CompareTo(Comparable) int
}
