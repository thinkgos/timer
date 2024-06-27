package timer

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Timer_Init(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		tm := NewTimer()
		require.Equal(t, int64(1), tm.TickMs())
		require.Equal(t, 32, tm.WheelSize())
		require.Equal(t, int64(0), tm.TaskCounter())
	})
	t.Run("custom", func(t *testing.T) {
		tm := NewTimer(WithTickMs(2), WithWheelSize(16), WithGoPool(goroutinePool))
		require.Equal(t, int64(2), tm.TickMs())
		require.Equal(t, 16, tm.WheelSize())
		require.Equal(t, int64(0), tm.TaskCounter())
	})

	t.Run("custom invalid setting", func(t *testing.T) {
		require.Panics(t, func() {
			_ = NewTimer(WithTickMs(-1))
		})
		require.Panics(t, func() {
			_ = NewTimer(WithWheelSize(-1))
		})
		require.NotPanics(t, func() {
			_ = NewTimer(WithGoPool(nil))
		})
	})
}

func Test_Timer_Start_Stop_Restart(t *testing.T) {
	tm := NewTimer()
	tm.Start()
	require.True(t, tm.Started())
	tm.Start() // double start, not start again.
	tm.Stop()
	require.False(t, tm.Started())
	time.Sleep(time.Millisecond * 100)
	tm.Start()
	require.True(t, tm.Started())
}

func ExampleTimer() {
	tm := NewTimer()
	tm.Start()
	tm.AfterFunc(100, func() {
		fmt.Println(100)
	})
	tm.AddTask(NewTask(200).WithJobFunc(func() {
		fmt.Println(200)
	}))
	canceledTaskAfterAdd := NewTask(300).WithJobFunc(func() {
		fmt.Println("canceled after add")
	})
	tm.AddTask(canceledTaskAfterAdd)
	canceledTaskAfterAdd.Cancel()
	canceledTaskBeforeAdd := NewTask(301).WithJobFunc(func() {
		fmt.Println("canceled before add")
	})
	canceledTaskBeforeAdd.Cancel()
	tm.AddTask(canceledTaskBeforeAdd)
	time.Sleep(time.Second)
	// Output:
	// 100
	// 200
}
