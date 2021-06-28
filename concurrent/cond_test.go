package concurrent

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type LockTestObject struct {
	t    *testing.T
	lock *sync.Mutex
	cond *TimeoutCond
}

func NewLockTestObject(t *testing.T) *LockTestObject {
	lock := new(sync.Mutex)
	return &LockTestObject{t: t, lock: lock, cond: NewTimeoutCond(lock)}
}

func (o *LockTestObject) lockAndWaitWithTimeout(timeout time.Duration) bool {
	o.lock.Lock()
	defer o.lock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return o.cond.Wait(ctx)
}

func (o *LockTestObject) lockAndWait() bool {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.t.Log("lockAndWait")
	return o.cond.Wait(context.Background())
}

func (o *LockTestObject) lockAndSignal() {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.t.Log("lockAndNotify")
	o.cond.Signal()
}

func (o *LockTestObject) hasWaiters() bool {
	return o.cond.HasWaiters()
}

func TestTimeoutCondWait(t *testing.T) {
	t.Parallel()

	t.Log("TestTimeoutCondWait")
	obj := NewLockTestObject(t)
	wait := sync.WaitGroup{}
	wait.Add(2)
	ch := make(chan bool, 1)
	go func() {
		ch <- obj.lockAndWait()
		wait.Done()
	}()
	time.Sleep(50 * time.Millisecond)
	go func() {
		obj.lockAndSignal()
		wait.Done()
	}()
	obj.lockAndSignal()
	wait.Wait()
	require.True(t, <-ch)
}

func TestTimeoutCondWaitTimeout(t *testing.T) {
	t.Parallel()

	t.Log("TestTimeoutCondWaitTimeout")
	obj := NewLockTestObject(t)
	wait := sync.WaitGroup{}
	wait.Add(1)
	ch := make(chan bool, 1)
	go func() {
		ch <- obj.lockAndWaitWithTimeout(2 * time.Second)
		wait.Done()
	}()
	wait.Wait()
	require.False(t, <-ch)
}

func TestTimeoutCondWaitTimeoutNotify(t *testing.T) {
	t.Parallel()

	t.Log("TestTimeoutCondWaitTimeoutNotify")
	obj := NewLockTestObject(t)
	wait := sync.WaitGroup{}
	wait.Add(2)
	ch := make(chan bool, 1)
	elapsedCh := make(chan time.Duration, 1)
	timeout := 5 * time.Second
	go func() {
		begin := time.Now()
		ch <- obj.lockAndWaitWithTimeout(timeout * time.Millisecond)
		elapsed := time.Since(begin)
		elapsedCh <- elapsed
		wait.Done()
	}()
	time.Sleep(200 * time.Millisecond)
	go func() {
		obj.lockAndSignal()
		wait.Done()
	}()
	wait.Wait()
	elapsed := <-elapsedCh
	close(elapsedCh)
	assert.True(t, elapsed < timeout)
	assert.True(t, elapsed >= 200*time.Millisecond)
	require.True(t, <-ch)
}

func TestTimeoutCondWaitTimeoutRemain(t *testing.T) {
	t.Parallel()

	t.Log("TestTimeoutCondWaitTimeoutRemain")
	obj := NewLockTestObject(t)
	wait := sync.WaitGroup{}
	wait.Add(2)
	ch := make(chan bool, 1)
	timeout := 2 * time.Second
	go func() {
		ch <- obj.lockAndWaitWithTimeout(timeout)
		wait.Done()
	}()
	time.Sleep(200 * time.Millisecond)
	go func() {
		obj.lockAndSignal()
		wait.Done()
	}()
	wait.Wait()
	assert.True(t, <-ch, "should not have been interrupted (timed out?)")
}

func TestTimeoutCondHasWaiters(t *testing.T) {
	t.Parallel()

	t.Log("TestTimeoutCondHasWaiters")
	obj := NewLockTestObject(t)
	waitersCount := 2
	ch := make(chan struct{}, waitersCount)
	for i := 0; i < 2; i++ {
		go func() {
			obj.lockAndWait()
			ch <- struct{}{}
		}()
	}
	time.Sleep(50 * time.Millisecond)
	assert.True(t, obj.hasWaiters(), "Should have waiters")

	obj.lockAndSignal()
	<-ch
	assert.True(t, obj.hasWaiters(), "Should still have waiters")

	obj.lockAndSignal()
	<-ch
	assert.False(t, obj.hasWaiters(), "Should no longer have waiters")
}

func TestTooManyWaiters(t *testing.T) {
	t.Parallel()

	obj := NewLockTestObject(t)
	obj.cond.hasWaiters = math.MaxUint64

	require.Panics(t, func() { obj.lockAndWait() })
}

func TestRemoveWaiterUsedIncorrectly(t *testing.T) {
	t.Parallel()

	cond := NewTimeoutCond(&sync.Mutex{})
	require.Panics(t, cond.removeWaiter)
}

func TestSignalNoWait(t *testing.T) {
	t.Parallel()

	obj := NewLockTestObject(t)
	obj.cond.Signal()
}
