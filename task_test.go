package timer

import (
	"sync"
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

func Test_Task_Activated(t *testing.T) {
	tm := NewTimer()
	tm.Start()
	task := NewTask(10 * time.Millisecond)
	require.False(t, task.Activated())
	err := tm.AddTask(task)
	require.Nil(t, err)
	require.True(t, task.Activated())
	time.Sleep(time.Millisecond * 50)
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

	// `newTaskEntry` samples `time.Now()` on its own and rounds the due time up to
	// the next whole millisecond (@task_entry.newTaskEntry), so the reported expiry
	// is the floor of the true due time at the earliest, and one millisecond later at
	// most when the two samples straddle a millisecond boundary. It can never be
	// earlier.
	wantExpiryMs := expiryAt.UnixMilli()
	wantExpiryAt := time.UnixMilli(wantExpiryMs)
	require.GreaterOrEqual(t, task.Expiry(), wantExpiryMs)
	require.LessOrEqual(t, task.Expiry(), wantExpiryMs+1)
	require.GreaterOrEqual(t, task.ExpiryAt(), wantExpiryAt)
	require.LessOrEqual(t, task.ExpiryAt(), wantExpiryAt.Add(time.Millisecond))

	time.Sleep(time.Millisecond * 20)
	require.Equal(t, int64(-1), task.Expiry())
	require.True(t, task.ExpiryAt().IsZero())
}

// Test_TaskEntry_MaxDelay_NoOverflow guards the expiration arithmetic against
// int64 overflow: a delay at the top of `time.Duration` (≈292 years) must still
// produce a large positive expiration, not a negative one, which `TimingWheel.add`
// would read as already-expired and fire immediately.
func Test_TaskEntry_MaxDelay_NoOverflow(t *testing.T) {
	tm := NewTimer(WithTickMs(1), WithWheelSize(8))
	tm.Start()
	defer tm.Stop()

	var fired atomic.Int64
	task := NewTask(time.Duration(1<<63 - 1)).WithJobFunc(func() { fired.Add(1) })
	require.NoError(t, tm.AddTask(task))

	require.Greater(t, task.Expiry(), time.Now().UnixMilli(),
		"a max-duration task must expire far in the future, not in the past")
	time.Sleep(20 * time.Millisecond)
	require.Zero(t, fired.Load(), "a max-duration task must not fire immediately")
}

// quickPeriodic re-schedules every millisecond, bypassing the one-second floor of
// [timer.Periodic] so the cancellation test below can run fast.
type quickPeriodic struct {
	job Job
}

func (q quickPeriodic) Run() { q.job.Run() }

func (q quickPeriodic) NextDelay() time.Duration { return time.Millisecond }

// Test_PeriodicTask_Cancel verifies that `Task.Cancel` stops a recurring task for
// good: the re-schedule path (@Timer.addTaskEntry) checks the task still owns its
// entry before re-adding, so a cancelled task is never re-inserted after the
// hand-off and does not fire again.
func Test_PeriodicTask_Cancel(t *testing.T) {
	tm := NewTimer(WithTickMs(1), WithWheelSize(4))
	tm.Start()
	defer tm.Stop()

	// Hold the first cycle open inside its job, so `Cancel` deterministically lands
	// in the hand-off window: the entry has been flushed, the job is running, and
	// the re-schedule defer has not run yet. That is exactly the window where an
	// unconditional re-add would resurrect the task.
	entered := make(chan struct{})
	release := make(chan struct{})
	var first sync.Once
	fired := &atomic.Int64{}
	task := NewScheduleTask(time.Millisecond, quickPeriodic{job: JobFunc(func() {
		fired.Add(1)
		first.Do(func() { close(entered) })
		<-release
	})})
	require.NoError(t, tm.AddTask(task))

	<-entered // the first cycle is inside its job, before the re-schedule defer.
	task.Cancel()
	close(release) // let that cycle finish; it must not re-schedule.

	time.Sleep(20 * time.Millisecond) // several periods.
	require.Equal(t, int64(1), fired.Load(), "periodic task kept firing after Cancel")
}
