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
	require.True(t, IsPowOf2(NextPowOf2(math.MaxInt64-200)))
}
