package timer

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testWantJobValue int64 = 6666

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
	t.val.Store(testWantJobValue)
}

func Test_Task_Job(t *testing.T) {
	job := newTestJob(100)
	task := NewTaskJob(100*time.Millisecond, job)
	require.Equal(t, 100*time.Millisecond, task.Delay())
	require.Equal(t, time.Second, task.SetDelay(time.Second).Delay())
	task.Run()
	require.Equal(t, testWantJobValue, job.Value())

	job1 := newTestJob(101)
	task1 := NewTaskFunc(101*time.Millisecond, job1.Run)
	task1.Run()
	require.Equal(t, testWantJobValue, job1.Value())

	// empty job
	task2 := NewTask(101 * time.Millisecond)
	task2.Run()
}

func Test_Task_RecoverPanic(t *testing.T) {
	task := NewTaskFunc(100*time.Millisecond, func() {
		panic("panic")
	})
	require.NotPanics(t, task.Run)
}

func Test_Task_Activated(t *testing.T) {
	tm := NewTimer()
	tm.Start()
	task := NewTask(10 * time.Millisecond)
	require.False(t, task.Activated())
	err := tm.AddTask(task)
	require.Nil(t, err)
	require.True(t, task.Activated())
	time.Sleep(time.Millisecond * 20)
	require.False(t, task.Activated())
}

func Test_Task_Expiry(t *testing.T) {
	delay := 10 * time.Millisecond

	tm := NewTimer()
	tm.Start()
	task := NewTask(delay)
	require.Equal(t, int64(-1), task.Expiry())
	require.True(t, task.ExpiryAt().IsZero())

	expiryAt := time.Now().Add(delay)
	err := tm.AddTask(task)
	require.Nil(t, err)

	wantExpiryMs := expiryAt.UnixMilli()
	wantExpiryAt := time.UnixMilli(wantExpiryMs)
	require.Equal(t, wantExpiryMs, task.Expiry())
	require.Equal(t, wantExpiryAt, task.ExpiryAt())

	time.Sleep(time.Millisecond * 20)
	require.Equal(t, int64(-1), task.Expiry())
	require.True(t, task.ExpiryAt().IsZero())
}
