package timer

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_DefaultTimer(t *testing.T) {
	require.NotNil(t, DefaultTimer())
	require.Equal(t, int64(DefaultTickMs), TickMs())
	require.Equal(t, DefaultWheelSize, WheelSize())
	require.GreaterOrEqual(t, TaskCounter(), int64(0))
}

func ExampleDefaultTimer() {
	fmt.Println(Started())
	Start()
	_, _ = AfterFunc(100*time.Millisecond, func() {
		fmt.Println(100)
	})
	canceledTaskThenAddAgain := NewTask(1100 * time.Millisecond).WithJobFunc(func() {
		fmt.Println("canceled then add again")
	})
	_ = AddTask(canceledTaskThenAddAgain)
	canceledTaskThenAddAgain.Cancel()
	_ = AddTask(NewTask(1025 * time.Millisecond).WithJobFunc(func() {
		fmt.Println(200)
	}))
	_ = AddTask(canceledTaskThenAddAgain)
	time.Sleep(time.Second + time.Millisecond*200)
	Stop()
	// Output:
	// true
	// 100
	// 200
	// canceled then add again
}
