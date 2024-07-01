package comparator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Comparable_Time(t *testing.T) {
	now := time.Now()
	v1 := now.Add(time.Minute)
	v2 := now
	require.Equal(t, 1, CompareTime(v1, v2))
	require.Equal(t, 0, CompareTime(v1, v1))
	require.Equal(t, -1, CompareTime(v2, v1))
}
