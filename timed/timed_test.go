package timed

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/thinkgos/timer"
)

func Test_Timed(t *testing.T) {
	require.Equal(t, int64(timer.DefaultTickMs), TickMs())
	require.Equal(t, timer.DefaultWheelSize, WheelSize())
	require.GreaterOrEqual(t, TaskCounter(), int64(0))
}

func ExampleTimer() {
	fmt.Println(Started())
	Start()
	_, _ = AfterFunc(100*time.Millisecond, func() {
		fmt.Println(100)
	})
	_ = AddTask(NewTask(1025 * time.Millisecond).WithJobFunc(func() {
		fmt.Println(200)
	}))
	canceledTaskAfterAdd := NewTaskFunc(300*time.Millisecond, func() {
		fmt.Println("canceled after add")
	})
	_ = AddTask(canceledTaskAfterAdd)
	canceledTaskAfterAdd.Cancel()
	time.Sleep(time.Second + time.Millisecond*200)
	Stop()
	// Output:
	// true
	// 100
	// 200
}
