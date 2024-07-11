package timer

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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

func Test_Task_RecoverPanic(t *testing.T) {
	task := NewTaskFunc(100*time.Millisecond, func() {
		panic(errors.New("panic"))
	})
	require.NotPanics(t, task.Run)
}

func Test_Task_Activated(t *testing.T) {
	task := NewTask(100 * time.Millisecond)
	require.False(t, task.Activated())
}
