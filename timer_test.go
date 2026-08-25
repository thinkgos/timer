package timer

import (
	"fmt"
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
