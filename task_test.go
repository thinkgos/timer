package timer

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	wantJobValue int64 = 6666
)

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
	task := NewTask(100).WithJob(job)
	require.Equal(t, int64(100), task.DelayMs())
	task.Run()
	require.Equal(t, wantJobValue, job.Value())

	job1 := newTestJob(101)
	task1 := NewTaskFunc(101, job1.Run)
	task1.Run()
	require.Equal(t, wantJobValue, job1.Value())

	// empty job
	task2 := NewTask(101)
	task2.Run()
}

func Test_Task_RecoverPanic(t *testing.T) {
	task := NewTaskFunc(100, func() {
		panic(errors.New("panic"))
	})
	require.NotPanics(t, task.Run)
}

func Test_Task_Cancel(t *testing.T) {
	task := NewTask(100)
	require.False(t, task.cancelled())

	task.Cancel()
	require.True(t, task.cancelled())
}
