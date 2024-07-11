package timed

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/thinkgos/timer"
)

const wantJobValue int64 = 6666

type testJob struct {
	val atomic.Int64
}

func newTestJob(val int64) *testJob {
	t := &testJob{}
	t.val.Store(val)
	return t
}
func (t *testJob) Value() int64 {
	return t.val.Load()
}
func (t *testJob) Run() {
	t.val.Store(wantJobValue)
}

func Test_Task_Job(t *testing.T) {
	job := newTestJob(100)
	task := NewTaskJob(100*time.Millisecond, job)
	require.Equal(t, 100*time.Millisecond, task.Delay())
	task.Run()
	require.Equal(t, wantJobValue, job.Value())

	job1 := newTestJob(101)
	task1 := NewTaskFunc(101*time.Millisecond, job1.Run)
	task1.Run()
	require.Equal(t, wantJobValue, job1.Value())

	// empty job
	task2 := NewTask(101 * time.Millisecond)
	task2.Run()
}
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
