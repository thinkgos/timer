package timer

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
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
		require.Panics(t, func() {
			_ = NewTimer(WithWheelSize(0))
		})
		require.Panics(t, func() {
			// no power of 2 fits, `NextPowOf2` reports 0 rather than overflowing.
			_ = NewTimer(WithWheelSize(math.MaxInt))
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
	err = tm.AddDerefTask(NewTask(100 * time.Millisecond))
	require.Nil(t, err)

	tm.Start() // double start, not start again.
	tm.Stop()
	require.False(t, tm.Started())
	time.Sleep(time.Millisecond * 100)
	tm.Start()
	require.True(t, tm.Started())
}

// Test_Timer_ConcurrentStartStop guards the start/stop handshake. `Stop` must wait
// outside `Timer.rw`, because the goroutine needs that lock to finish its current
// tick, and that window lets a concurrent `Start` begin a new run. Tracking runs with
// a single `sync.WaitGroup` breaks there two ways: `Stop` waits on the run `Start`
// just added and blocks forever, or the `Add` races the `Wait` and panics with
// "sync: WaitGroup misuse: Add called concurrently with Wait". see `Timer.Stop`.
func Test_Timer_ConcurrentStartStop(t *testing.T) {
	const (
		rounds      = 300
		workerCount = 4
	)

	for i := 0; i < rounds; i++ {
		tm := NewTimer(WithTickMs(1), WithWheelSize(4))
		tm.Start()
		require.NoError(t, tm.AddTask(NewPeriodicTaskFunc(time.Second, func() {})))

		// release all workers at once, to land inside the unlocked window of `Stop`.
		begin := make(chan struct{})
		settled := make(chan struct{})
		go func() {
			defer close(settled)
			wg := sync.WaitGroup{}
			for w := 0; w < workerCount; w++ {
				wg.Add(1)
				go func(w int) {
					defer wg.Done()
					<-begin
					if w%2 == 0 {
						tm.Stop()
					} else {
						tm.Start()
					}
				}(w)
			}
			wg.Wait()
			tm.Stop()
		}()

		close(begin)
		select {
		case <-settled:
		case <-time.After(10 * time.Second):
			t.Fatalf("round %d: Stop did not return, it is waiting on a run that Start replaced", i)
		}
		require.False(t, tm.Started())
	}
}

// Test_Timer_StopIdempotent covers `Stop` before any start (`Timer.done` is still
// nil) and repeated `Stop` on an already stopped timer.
func Test_Timer_StopIdempotent(t *testing.T) {
	tm := NewTimer()
	require.False(t, tm.Started())
	tm.Stop() // never started, must not block nor panic.
	tm.Stop()

	tm.Start()
	require.True(t, tm.Started())
	tm.Stop()
	tm.Stop()
	require.False(t, tm.Started())

	// a stopped timer can be restarted and stopped again.
	tm.Start()
	require.True(t, tm.Started())
	require.NoError(t, tm.AddTask(NewTask(time.Millisecond)))
	tm.Stop()
	require.False(t, tm.Started())
}

func Test_Timer_InvalidDelay(t *testing.T) {
	tm := NewTimer()
	tm.Start()
	defer tm.Stop()

	t.Run("negative", func(t *testing.T) {
		require.ErrorIs(t, tm.AddTask(NewTask(-time.Second)), ErrInvalidDelay)
		require.ErrorIs(t, tm.AddDerefTask(NewTask(-time.Second)), ErrInvalidDelay)
		task, err := tm.AfterFunc(-time.Second, func() {})
		require.ErrorIs(t, err, ErrInvalidDelay)
		require.Nil(t, task)
	})
	t.Run("zero", func(t *testing.T) {
		require.ErrorIs(t, tm.AddTask(NewTask(0)), ErrInvalidDelay)
		_, err := tm.AfterFunc(0, func() {})
		require.ErrorIs(t, err, ErrInvalidDelay)
	})
	t.Run("exhausted schedule", func(t *testing.T) {
		// a crontab whose expression has no next run reports a delay of -1.
		task, err := NewCrontabTask("0 0 1 1 * 2000", JobFunc(func() {}))
		require.NoError(t, err)
		require.ErrorIs(t, tm.AddTask(task), ErrInvalidDelay)
	})
	t.Run("rejected task is not activated", func(t *testing.T) {
		task := NewTask(-time.Second)
		require.ErrorIs(t, tm.AddTask(task), ErrInvalidDelay)
		require.False(t, task.Activated())
		require.Equal(t, int64(-1), task.Expiry())
	})
}

// Test_Timer_ConcurrentCancelAndAdd guards the lock order between `Task.rw` and
// `Spoke.mu`. `Task.Cancel` takes `Task.rw` then `Spoke.mu`, while the timer
// goroutine's `Spoke.Flush` -> `Timer.addTaskEntry` -> `taskEntry.cancelled` takes
// `Task.rw`. Applying the flush function under `Spoke.mu` inverts the order and
// deadlocks the timer goroutine. see `Spoke.Flush`.
func Test_Timer_ConcurrentCancelAndAdd(t *testing.T) {
	const (
		taskCount   = 200
		workerCount = 8
		stressFor   = 2 * time.Second
	)

	tm := NewTimer(WithTickMs(1), WithWheelSize(4))
	tm.Start()

	fired := &atomic.Int64{}
	tasks := make([]*Task, taskCount)
	for i := range tasks {
		tasks[i] = NewPeriodicTaskFunc(time.Second, func() { fired.Add(1) })
		tasks[i].SetDelay(time.Millisecond)
		require.NoError(t, tm.AddTask(tasks[i]))
	}

	done := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		wg := sync.WaitGroup{}
		for w := 0; w < workerCount; w++ {
			wg.Add(1)
			go func(w int) {
				defer wg.Done()
				for {
					select {
					case <-done:
						return
					default:
					}
					for i := w; i < taskCount; i += workerCount {
						tasks[i].Cancel()
						tasks[i].SetDelay(time.Millisecond)
						_ = tm.AddTask(tasks[i])
					}
					// Pace the workers: without this they re-arm every task's 1ms
					// delay faster than it can elapse, so no task ever fires and the
					// `fired` assertion below is a coin flip (~93% fail on this box).
					time.Sleep(200 * time.Microsecond)
				}
			}(w)
		}
		wg.Wait()
	}()

	time.Sleep(stressFor)
	close(done)
	select {
	case <-finished:
	case <-time.After(10 * time.Second):
		t.Fatal("deadlock: Task.Cancel and Spoke.Flush deadlocked on Task.rw/Spoke.mu")
	}

	require.NotZero(t, fired.Load())
	// every added task entry is eventually removed exactly once, so once the
	// in-flight entries are flushed the counter settles back to the live task count.
	require.Eventually(t, func() bool {
		return tm.TaskCounter() == taskCount
	}, 5*time.Second, 50*time.Millisecond)
	tm.Stop()
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
