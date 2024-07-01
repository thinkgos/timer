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

	t := NewTask(100 * time.Millisecond).WithJobFunc(func() {
		fmt.Println(100)
	})
	_ = AddTask(t)
	_, _ = AfterFunc(1025*time.Millisecond, func() {
		fmt.Println(200)
	})

	canceledTaskAfterAdd := NewTaskFunc(300*time.Millisecond, func() {
		fmt.Println("canceled after add")
	})
	_ = AddTask(canceledTaskAfterAdd)
	canceledTaskAfterAdd.Cancel()
	canceledTaskBeforeAdd := NewTask(301 * time.Millisecond).WithJobFunc(func() {
		fmt.Println("canceled before add")
	})
	canceledTaskBeforeAdd.Cancel()
	_ = AddTask(canceledTaskBeforeAdd)
	time.Sleep(time.Millisecond * 500)
	_ = AddTask(t.Reset())
	time.Sleep(time.Second + time.Millisecond*200)
	Stop()
	// Output:
	// true
	// 100
	// 100
	// 200
}
