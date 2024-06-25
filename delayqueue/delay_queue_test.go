package delayqueue

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/things-go/timer/queue"
)

func nowMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

type delay struct {
	name  string
	value int64
}

func (sf delay) Value() int64 {
	return atomic.LoadInt64(&sf.value)
}

func (sf delay) DelayMs() int64 {
	return atomic.LoadInt64(&sf.value) - nowMs()
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

func CompareDelay(v1, v2 interface{}) int {
	vv1 := v1.(*delay)
	vv2 := v2.(*delay)

	if vv1.Value() < vv2.Value() {
		return -1
	}
	if vv1.Value() > vv2.Value() {
		return 1
	}
	return 0
}

func TestDelayQueue(t *testing.T) {
	dq := NewDelayQueue[*delay]()

	d1 := &delay{"d1", nowMs() + 1000}
	d2 := &delay{"d2", nowMs() + 2000}
	dq.Add(d1)
	dq.Add(d2)
	v1 := dq.Take(context.Background())
	assert.Equal(t, "d1", v1.(*delay).name)
	assert.LessOrEqual(t, v1.DelayMs(), int64(0))

	v2 := dq.Take(context.Background())
	assert.Equal(t, "d2", v2.(*delay).name)
	assert.LessOrEqual(t, v2.DelayMs(), int64(0))
}

func TestDelayQueue_Empty_Begin(t *testing.T) {
	dq := NewDelayQueue[*delay]()

	go func() {
		time.Sleep(time.Millisecond * 200)
		d1 := &delay{"d1", nowMs() + 1000}
		d2 := &delay{"d2", nowMs() + 2000}
		dq.Add(d1)
		dq.Add(d2)
	}()
	v1 := dq.Take(context.Background())
	assert.Equal(t, "d1", v1.(*delay).name)
	assert.LessOrEqual(t, v1.DelayMs(), int64(0))

	v2 := dq.Take(context.Background())
	assert.Equal(t, "d2", v2.(*delay).name)
	assert.LessOrEqual(t, v2.DelayMs(), int64(0))
}

func TestDelayQueue_Cancel(t *testing.T) {
	dq := NewDelayQueue[*delay]()

	d1 := &delay{"d1", nowMs() + 1000}
	d2 := &delay{"d2", nowMs() + 2000}
	dq.Add(d1)
	dq.Add(d2)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()
	vxx := dq.Take(ctx)
	assert.Nil(t, vxx)

	v1 := dq.Take(context.Background())
	assert.Equal(t, "d1", v1.(*delay).name)
	assert.LessOrEqual(t, v1.DelayMs(), int64(0))

	v2 := dq.Take(context.Background())
	assert.Equal(t, "d2", v2.(*delay).name)
	assert.LessOrEqual(t, v2.DelayMs(), int64(0))
}

func TestDelayQueue_Cancel_Empty_Begin(t *testing.T) {
	dq := NewDelayQueue[*delay]()

	go func() {
		time.Sleep(time.Millisecond * 500)
		d1 := &delay{"d1", nowMs() + 1000}
		d2 := &delay{"d2", nowMs() + 2000}
		dq.Add(d1)
		dq.Add(d2)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()
	vxx := dq.Take(ctx)
	assert.Nil(t, vxx)

	v1 := dq.Take(context.Background())
	assert.Equal(t, "d1", v1.(*delay).name)
	assert.LessOrEqual(t, v1.DelayMs(), int64(0))

	v2 := dq.Take(context.Background())
	assert.Equal(t, "d2", v2.(*delay).name)
	assert.LessOrEqual(t, v2.DelayMs(), int64(0))
}
