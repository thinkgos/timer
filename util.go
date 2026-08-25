package timer

import "math"

// IsPowOf2 reports whether x is a power of 2.
// Only positive x can be a power of 2, so 0 and negative values report false.
func IsPowOf2(x int) bool {
	return x > 0 && (x&(x-1)) == 0
}

// NextPowOf2 return the least power of 2 that is greater than or equal to x.
// It returns 0 when x is not positive, or when the next power of 2 does not fit
// in an int, so that a caller validating the result rejects the input. see `NewTimer`.
func NextPowOf2(x int) int {
	if x <= 0 {
		return 0
	}
	if IsPowOf2(x) {
		return x
	}
	// smear the highest set bit down over every lower bit, then carry once.
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x |= x >> 32 // int is 64 bits wide here; on a 32 bits platform this shift is a no-op.
	if x == math.MaxInt {
		return 0 // x is above the largest representable power of 2, `x+1` would overflow.
	}
	return x + 1
}
