package delayqueue

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thinkgos/timer/queue"
)

type delay struct {
	name  string
	value int64
}

func (d delay) Value() int64 {
	return atomic.LoadInt64(&d.value)
}

func (d delay) DelayMs() int64 {
	return atomic.LoadInt64(&d.value) - time.Now().UnixMilli()
}

func (v1 *delay) CompareTo(v2 queue.Comparable) int {
	vv2 := v2.(*delay)

	if v1.Value() < vv2.Value() {
		return -1
	}
	if v1.Value() > vv2.Value() {
		return 1
	}
	return 0
}

func Test_DelayQueue(t *testing.T) {
	dq := NewDelayQueue[*delay]()

	d1 := &delay{"d1", time.Now().UnixMilli() + 100}
	d2 := &delay{"d2", time.Now().UnixMilli() + 200}
	dq.Add(d1)
	dq.Add(d2)
	v1, exit := dq.Take(nil)
	require.False(t, exit)
	assert.Equal(t, "d1", v1.name)
	assert.LessOrEqual(t, v1.DelayMs(), int64(0))

	v2, exit := dq.Take(nil)
	require.False(t, exit)
	assert.Equal(t, "d2", v2.name)
	assert.LessOrEqual(t, v2.DelayMs(), int64(0))
}

func Test_DelayQueue_Empty_Begin(t *testing.T) {
	dq := NewDelayQueue[*delay]()

	go func() {
		time.Sleep(time.Millisecond * 20)
		d1 := &delay{"d1", time.Now().UnixMilli() + 100}
		dq.Add(d1)
	}()
	go func() {
		time.Sleep(time.Millisecond * 20)
		d2 := &delay{"d2", time.Now().UnixMilli() + 200}
		dq.Add(d2)
	}()
	go func() {
		time.Sleep(time.Millisecond * 20)
		d3 := &delay{"d3", time.Now().UnixMilli() + 50}
		dq.Add(d3)
	}()

	v1, exit := dq.Take(nil)
	require.False(t, exit)
	assert.Equal(t, "d3", v1.name)
	assert.LessOrEqual(t, v1.DelayMs(), int64(0))

	v3, exit := dq.Take(nil)
	require.False(t, exit)
	assert.Equal(t, "d1", v3.name)
	assert.LessOrEqual(t, v3.DelayMs(), int64(0))

	v2, exit := dq.Take(nil)
	require.False(t, exit)
	assert.Equal(t, "d2", v2.name)
	assert.LessOrEqual(t, v2.DelayMs(), int64(0))
}

func Test_DelayQueue_Quit(t *testing.T) {
	dq := NewDelayQueue[*delay]()

	d1 := &delay{"d1", time.Now().UnixMilli() + 100}
	d2 := &delay{"d2", time.Now().UnixMilli() + 200}
	dq.Add(d1)
	dq.Add(d2)

	quitChan := make(chan struct{})
	close(quitChan)
	vxx, exit := dq.Take(quitChan)
	require.True(t, exit)
	assert.Nil(t, vxx)

	v1, exit := dq.Take(nil)
	require.False(t, exit)
	assert.Equal(t, "d1", v1.name)
	assert.LessOrEqual(t, v1.DelayMs(), int64(0))

	v2, exit := dq.Take(nil)
	require.False(t, exit)
	assert.Equal(t, "d2", v2.name)
	assert.LessOrEqual(t, v2.DelayMs(), int64(0))
}

func Test_DelayQueue_Quit_Empty_Begin(t *testing.T) {
	dq := NewDelayQueue[*delay]()

	go func() {
		time.Sleep(time.Millisecond * 50)
		d1 := &delay{"d1", time.Now().UnixMilli() + 100}
		d2 := &delay{"d2", time.Now().UnixMilli() + 200}
		dq.Add(d1)
		dq.Add(d2)
	}()

	quitChan := make(chan struct{})
	close(quitChan)
	vxx, exit := dq.Take(quitChan)
	require.True(t, exit)
	assert.Nil(t, vxx)

	v1, exit := dq.Take(nil)
	require.False(t, exit)
	assert.Equal(t, "d1", v1.name)
	assert.LessOrEqual(t, v1.DelayMs(), int64(0))

	v2, exit := dq.Take(nil)
	require.False(t, exit)
	assert.Equal(t, "d2", v2.name)
	assert.LessOrEqual(t, v2.DelayMs(), int64(0))
}
