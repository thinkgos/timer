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
	require.Zero(t, spoke1.DelayMs())
	spoke2 := NewSpoke(&atomic.Int64{})
	require.Equal(t, int64(-1), spoke2.GetExpiration())
	require.Zero(t, spoke2.DelayMs())

	now := time.Now()
	require.True(t, spoke1.SetExpiration(now.Add(time.Minute*2).UnixMilli()))
	require.True(t, spoke2.SetExpiration(now.Add(time.Minute).UnixMilli()))
	require.NotZero(t, spoke1.DelayMs())

	require.Equal(t, 0, spoke1.CompareTo(spoke1))
	require.Equal(t, 1, spoke1.CompareTo(spoke2))
	require.Equal(t, -1, spoke2.CompareTo(spoke1))
}

func Test_Spoke_Task(t *testing.T) {
	tasks := map[*Task]struct{}{
		NewTask(101): {},
		NewTask(102): {},
		NewTask(103): {},
		NewTask(105): {},
	}
	task1 := NewTask(104)

	taskCounter := &atomic.Int64{}
	spoke := NewSpoke(taskCounter)
	spoke.Add(task1)
	for task := range tasks {
		spoke.Add(task)
	}
	require.Equal(t, int64(5), taskCounter.Load())
	spoke.Remove(task1)
	require.Equal(t, int64(4), taskCounter.Load())

	spoke.Flush(func(task *Task) {
		_, ok := tasks[task]
		require.True(t, ok)
		delete(tasks, task)
	})
	require.Equal(t, int64(0), taskCounter.Load())
}
