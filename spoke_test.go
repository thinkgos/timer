package timer

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Spoke(t *testing.T) {
	spoke1 := NewSpoke(&atomic.Int64{})
	require.Equal(t, int64(-1), spoke1.GetExpiration())
	require.Zero(t, spoke1.Delay())
	spoke2 := NewSpoke(&atomic.Int64{})
	require.Equal(t, int64(-1), spoke2.GetExpiration())
	require.Zero(t, spoke2.Delay())

	now := time.Now()
	require.True(t, spoke1.SetExpiration(now.Add(time.Minute*2).UnixMilli()))
	require.True(t, spoke2.SetExpiration(now.Add(time.Minute).UnixMilli()))
	require.NotZero(t, spoke1.Delay())

	require.Equal(t, 0, CompareSpoke(spoke1, spoke1))
	require.Equal(t, 1, CompareSpoke(spoke1, spoke2))
	require.Equal(t, -1, CompareSpoke(spoke2, spoke1))
}

func Test_Spoke_Task(t *testing.T) {
	tasks := map[*taskEntry]struct{}{
		newTaskEntry(NewTask(101 * time.Millisecond)): {},
		newTaskEntry(NewTask(102 * time.Millisecond)): {},
		newTaskEntry(NewTask(103 * time.Millisecond)): {},
		newTaskEntry(NewTask(105 * time.Millisecond)): {},
	}
	task1 := newTaskEntry(NewTask(104 * time.Millisecond))

	taskCounter := &atomic.Int64{}
	spoke := NewSpoke(taskCounter)
	spoke.Add(task1)
	for task := range tasks {
		spoke.Add(task)
	}
	require.Equal(t, int64(5), taskCounter.Load())
	spoke.Remove(task1)
	require.Equal(t, int64(4), taskCounter.Load())

	spoke.Flush(func(task *taskEntry) {
		_, ok := tasks[task]
		require.True(t, ok)
		delete(tasks, task)
	})
	require.Equal(t, int64(0), taskCounter.Load())
}

// Test_Spoke_Flush_ReinsertDuringWalk guards the chain walk in `Spoke.Flush`. The
// detached chain is the work list, so the walk must read an entry's successor BEFORE
// applying the function: the function re-inserts the entry (@Timer.addTaskEntry ->
// @Spoke.Add), and `Spoke.pushBack` overwrites the entry's links. Reading the
// successor afterwards walks into the spoke the entry was just re-inserted into,
// which truncates the flush. Re-inserting into the very spoke being flushed is the
// worst case, and the one this test pins down.
func Test_Spoke_Flush_ReinsertDuringWalk(t *testing.T) {
	const taskCount = 64

	taskCounter := &atomic.Int64{}
	spoke := NewSpoke(taskCounter)
	want := make([]*taskEntry, 0, taskCount)
	for i := 0; i < taskCount; i++ {
		te := newTaskEntry(NewTask(time.Duration(100+i) * time.Millisecond))
		want = append(want, te)
		spoke.Add(te)
	}
	require.Equal(t, int64(taskCount), taskCounter.Load())

	got := make([]*taskEntry, 0, taskCount)
	spoke.Flush(func(te *taskEntry) {
		got = append(got, te)
		spoke.Add(te) // re-insert into the spoke being flushed.
	})

	// every entry visited exactly once, in insertion order, and all of them landed
	// back on the spoke rather than being lost with a truncated chain.
	require.Equal(t, want, got)
	require.Equal(t, int64(taskCount), taskCounter.Load())
}

// Test_Spoke_Flush_ApplyOutOfLock asserts `Spoke.Flush` releases the spoke's lock
// before applying the supplied function. Applying it under the lock inverts the
// `Task.rw` -> `Spoke.mu` order taken by `Task.Cancel` and deadlocks. see `Spoke.Flush`.
func Test_Spoke_Flush_ApplyOutOfLock(t *testing.T) {
	spoke := NewSpoke(&atomic.Int64{})
	spoke.Add(newTaskEntry(NewTask(101 * time.Millisecond)))
	spoke.Add(newTaskEntry(NewTask(102 * time.Millisecond)))

	spoke.Flush(func(te *taskEntry) {
		unlocked := make(chan struct{})
		go func() {
			defer close(unlocked)
			spoke.Remove(te) // contends on `Spoke.mu`, as `Task.Cancel` would.
		}()
		select {
		case <-unlocked:
		case <-time.After(time.Second):
			t.Error("Spoke.Flush still holds the lock while applying the function")
		}
	})
}

func Benchmark_Spoke_Flush(b *testing.B) {
	for _, taskCount := range []int{1, 16, 256, 4096} {
		b.Run(fmt.Sprintf("tasks-%d", taskCount), func(b *testing.B) {
			spoke := NewSpoke(&atomic.Int64{})
			entries := make([]*taskEntry, taskCount)
			for i := range entries {
				entries[i] = newTaskEntry(NewTask(time.Second))
			}
			noop := func(*taskEntry) {}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				for _, te := range entries {
					spoke.Add(te)
				}
				b.StartTimer()
				spoke.Flush(noop)
			}
		})
	}
}
