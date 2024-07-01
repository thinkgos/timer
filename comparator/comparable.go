package comparator

import "time"

// Comparable method CompareTo(Comparable) should return a negative number when v1 < v2,
// a positive number when v1 > v2 and zero when v1 == v2.
type Comparable[T any] func(T, T) int

func CompareTime(v1, v2 time.Time) int {
	t1 := time.Time(v1)
	t2 := time.Time(v2)
	if t1.Before(t2) {
		return -1
	}
	if t1.After(t2) {
		return 1
	}
	return 0
}
