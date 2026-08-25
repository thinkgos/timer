package timer

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func ranges(start, end int) []int {
	length := int(math.Abs(float64(end-start))) + 1
	res := make([]int, 0, length)
	for i := 0; i < length; i++ {
		value := start
		if end > start {
			value += i
		} else {
			value -= i
		}
		res = append(res, value)
	}
	return res
}

func Test_IsPowOf2(t *testing.T) {
	for _, v := range ranges(0, 62) {
		vv := int(math.Pow(2, float64(v)))
		require.True(t, IsPowOf2(vv), "%d - %d", vv, v)
	}
	require.False(t, IsPowOf2(100))
	// only a positive number can be a power of 2.
	require.False(t, IsPowOf2(0))
	require.False(t, IsPowOf2(-1))
	require.False(t, IsPowOf2(-2))
	require.False(t, IsPowOf2(math.MinInt)) // -2^63, `x & (x-1)` alone reports it as one.
}

func Test_NextPowOf2(t *testing.T) {
	for _, v := range ranges(3, 62) {
		want := int(math.Pow(2, float64(v)))
		got := NextPowOf2(want - 2)
		require.Equal(t, want, got)
		require.True(t, IsPowOf2(got))

		got2 := NextPowOf2(want)
		require.Equal(t, want, got2)
		require.True(t, IsPowOf2(got2))
	}
	// a single set bit plus 1: the lower bits are not already set, so the smear must
	// reach all the way down. Without the `>> 32` step this is wrong above 2^32.
	for _, v := range ranges(1, 61) {
		x := int(math.Pow(2, float64(v))) + 1
		want := x - 1 + int(math.Pow(2, float64(v)))
		got := NextPowOf2(x)
		require.Equal(t, want, got, "2^%d+1", v)
		require.True(t, IsPowOf2(got), "2^%d+1 -> %d", v, got)
	}
	require.Equal(t, 1, NextPowOf2(1))
	require.Equal(t, 2, NextPowOf2(2))
	require.Equal(t, 4, NextPowOf2(3))
	// out of domain, and the overflow, are both reported as 0 so that `NewTimer` panics.
	require.Equal(t, 0, NextPowOf2(0))
	require.Equal(t, 0, NextPowOf2(-1))
	require.Equal(t, 0, NextPowOf2(math.MinInt))
	require.Equal(t, 0, NextPowOf2(math.MaxInt))
	require.Equal(t, 0, NextPowOf2(math.MaxInt-200))
	// the largest input that still fits.
	require.Equal(t, 1<<62, NextPowOf2(1<<62))
	require.Equal(t, 1<<62, NextPowOf2(1<<62-1))
	require.Equal(t, 0, NextPowOf2(1<<62+1))
}
