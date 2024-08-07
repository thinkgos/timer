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
		require.Equal(t, int64(DefaultTickMs), tm.TickMs())
		require.Equal(t, DefaultWheelSize, tm.WheelSize())
		require.Equal(t, int64(0), tm.TaskCounter())
	})
	t.Run("custom", func(t *testing.T) {
		tm := NewTimer(WithTickMs(2), WithWheelSize(16), WithGoPool(goroutinePool))
		require.Equal(t, int64(2), tm.TickMs())
		require.Equal(t, 16, tm.WheelSize())
		require.Equal(t, 0xf, tm.WheelMask())
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
	// timer is closed
	_, err := tm.AfterFunc(time.Second, func() {})
	require.ErrorIs(t, err, ErrClosed)
	err = tm.AddTask(NewTask(100 * time.Millisecond))
	require.ErrorIs(t, err, ErrClosed)
	tm.Start()
	require.True(t, tm.Started())
	// timer is started
	_, err = tm.AfterFunc(time.Millisecond*100, func() {})
	require.Nil(t, err)
	err = tm.AddTask(NewTask(100 * time.Millisecond))
	require.Nil(t, err)

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
	_, _ = tm.AfterFunc(100*time.Millisecond, func() {
		fmt.Println(100)
	})
	canceledTaskThenAddAgain := NewTask(1100 * time.Millisecond).WithJobFunc(func() {
		fmt.Println("canceled then add again")
	})
	_ = tm.AddTask(canceledTaskThenAddAgain)
	canceledTaskThenAddAgain.Cancel()
	_ = tm.AddTask(NewTask(1025 * time.Millisecond).WithJobFunc(func() {
		fmt.Println(200)
	}))
	_ = tm.AddTask(canceledTaskThenAddAgain)
	time.Sleep(time.Second + time.Millisecond*200)
	tm.Stop()
	// Output:
	// 100
	// 200
	// canceled then add again
}
