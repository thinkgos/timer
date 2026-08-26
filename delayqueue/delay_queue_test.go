package delayqueue

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type delay struct {
	name  string
	value int64
}

func (d delay) Value() int64 {
	return atomic.LoadInt64(&d.value)
}

func (d delay) Delay() int64 {
	return atomic.LoadInt64(&d.value) - time.Now().UnixMilli()
}

func compareDelay(v1 *delay, v2 *delay) int {
	if v1.Value() < v2.Value() {
		return -1
	}
	if v1.Value() > v2.Value() {
		return 1
	}
	return 0
}

func Test_DelayQueue(t *testing.T) {
	dq := NewDelayQueue(compareDelay)

	d1 := &delay{"d1", time.Now().UnixMilli() + 100}
	d2 := &delay{"d2", time.Now().UnixMilli() + 200}
	dq.Add(d1)
	dq.Add(d2)
	v1, exit := dq.Take(nil)
	require.False(t, exit)
	assert.Equal(t, "d1", v1.name)
	assert.LessOrEqual(t, v1.Delay(), int64(0))

	v2, exit := dq.Take(nil)
	require.False(t, exit)
	assert.Equal(t, "d2", v2.name)
	assert.LessOrEqual(t, v2.Delay(), int64(0))
}

func Test_DelayQueue_Empty_Begin(t *testing.T) {
	dq := NewDelayQueue(compareDelay)

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
	assert.LessOrEqual(t, v1.Delay(), int64(0))

	v3, exit := dq.Take(nil)
	require.False(t, exit)
	assert.Equal(t, "d1", v3.name)
	assert.LessOrEqual(t, v3.Delay(), int64(0))

	v2, exit := dq.Take(nil)
	require.False(t, exit)
	assert.Equal(t, "d2", v2.name)
	assert.LessOrEqual(t, v2.Delay(), int64(0))
}

func Test_DelayQueue_Quit(t *testing.T) {
	dq := NewDelayQueue(compareDelay)

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
	assert.LessOrEqual(t, v1.Delay(), int64(0))

	v2, exit := dq.Take(nil)
	require.False(t, exit)
	assert.Equal(t, "d2", v2.name)
	assert.LessOrEqual(t, v2.Delay(), int64(0))
}

func Test_DelayQueue_Quit_Empty_Begin(t *testing.T) {
	dq := NewDelayQueue(compareDelay).TimeUnit(time.Millisecond)

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
	assert.LessOrEqual(t, v1.Delay(), int64(0))

	v2, exit := dq.Take(nil)
	require.False(t, exit)
	assert.Equal(t, "d2", v2.name)
	assert.LessOrEqual(t, v2.Delay(), int64(0))
}

// Test_DelayQueue_Take_Reuse_Timer covers the multi-wait path of `Take`. Every `Add`
// of a new head wakes the waiter up, so a single `Take` call resets its timer many
// times. Whatever happens to the timer, `Take` must never hand out an element before
// its delay has elapsed, and must not miss the last deadline it waited on.
func Test_DelayQueue_Take_Reuse_Timer(t *testing.T) {
	const rounds = 20

	dq := NewDelayQueue(compareDelay)
	dq.Add(&delay{"tail", time.Now().UnixMilli() + 1000})

	go func() {
		// each new element is the earliest one, so each `Add` notifies the waiter and
		// sends `Take` around the loop again with a shorter deadline.
		for i := rounds; i > 0; i-- {
			time.Sleep(2 * time.Millisecond)
			dq.Add(&delay{fmt.Sprintf("head-%d", i), time.Now().UnixMilli() + int64(20*i)})
		}
	}()

	begin := time.Now()
	got, exit := dq.Take(nil)
	require.False(t, exit)
	require.Equal(t, "head-1", got.name)
	// not handed out early, and not overslept either.
	assert.LessOrEqual(t, got.Delay(), int64(0))
	assert.Less(t, time.Since(begin), time.Second)
}

func Test_DelayQueue_Poll(t *testing.T) {
	dq := NewDelayQueue(compareDelay)

	d1 := &delay{"d1", time.Now().UnixMilli()}
	d2 := &delay{"d2", time.Now().UnixMilli() + 200}
	dq.Add(d1)
	dq.Add(d2)

	v1, exist := dq.Poll()
	require.True(t, exist)
	assert.Equal(t, "d1", v1.name)
	assert.LessOrEqual(t, v1.Delay(), int64(0))

	v2, exist := dq.Poll()
	require.False(t, exist)
	assert.Nil(t, v2)
}
