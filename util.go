package timer

import (
	"time"
)

func NowMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func IsPowOf2(x int) bool {
	return (x & (x - 1)) == 0
}

func NextPowOf2(x int) int {
	if IsPowOf2(x) {
		return x
	}
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	return x + 1
}
