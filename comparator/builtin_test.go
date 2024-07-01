package comparator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/constraints"
)

type testCombine[T any] interface {
	constraints.Ordered
	Comparable[T]
}

func pb_Test_Builtin[T testCombine[T]](t *testing.T, v1, v2 T, want int) {
	require.Equal(t, want, v1.CompareTo(v2))
}

func Test_Builtin(t *testing.T) {
	pb_Test_Builtin(t, Int(1), Int(2), -1)
	pb_Test_Builtin(t, Int(1), Int(1), 0)
	pb_Test_Builtin(t, Int(2), Int(1), 1)
	pb_Test_Builtin(t, Int8(2), Int8(1), 1)
	pb_Test_Builtin(t, Int16(2), Int16(1), 1)
	pb_Test_Builtin(t, Int32(2), Int32(1), 1)
	pb_Test_Builtin(t, Int64(2), Int64(1), 1)
	pb_Test_Builtin(t, Uint(2), Uint(1), 1)
	pb_Test_Builtin(t, Uint8(2), Uint8(1), 1)
	pb_Test_Builtin(t, Uint16(2), Uint16(1), 1)
	pb_Test_Builtin(t, Uint32(2), Uint32(1), 1)
	pb_Test_Builtin(t, Uint64(2), Uint64(1), 1)
	pb_Test_Builtin(t, Float32(2), Float32(1), 1)
	pb_Test_Builtin(t, Float64(2), Float64(1), 1)
	pb_Test_Builtin(t, Uintptr(2), Uintptr(1), 1)
	pb_Test_Builtin(t, String("2"), String("1"), 1)
}

func Test_Builtin_Time(t *testing.T) {
	now := time.Now()
	v1 := Time(now.Add(time.Minute))
	v2 := Time(now)
	require.Equal(t, 1, v1.CompareTo(v2))
	require.Equal(t, 0, v1.CompareTo(v1))
	require.Equal(t, -1, v2.CompareTo(v1))
}
