package timer

import (
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
