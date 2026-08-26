package timer

import (
	"sync/atomic"
	"time"
)

// taskEntry is an element of a linked list, hold the task instance.
type taskEntry struct {
	// next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	prev         *taskEntry
	next         *taskEntry
	list         atomic.Pointer[Spoke] // The list to which this element belongs.
	expirationMs int64                 // expiration time, absolute time(immutable after first initialization), Units: ms
	task         *Task                 // the task instance.
}

func newTaskEntry(task *Task) *taskEntry {
	now := time.Now()
	nowMs := now.UnixMilli()
	nowFrac := now.UnixNano() - nowMs*1e6
	delay := task.Delay()
	delayMs := delay.Milliseconds()
	delayFrac := delay.Nanoseconds() - delayMs*1e6
	expirationMs := nowMs + delayMs + (nowFrac+delayFrac+1e6-1)/1e6
	te := &taskEntry{
		task: task,
		// expirationMs returns the due time (now + delay) rounded up to the next whole
		// millisecond, so a task never fires before its delay has fully elapsed.
		// Truncating instead made a fractional delay fire up to one tick early, and a
		// sub-millisecond delay expire in the same millisecond it was added, so
		// `TimingWheel.add` judged it already expired and ran it immediately.
		//
		// the absolute expiration, in Unix milliseconds, of a task
		// scheduled `delay` from now, rounded up to the next whole millisecond.
		// It is deliberately not computed as a single (time.Now().UnixNano() + delay)/ms:
		// for a delay beyond roughly 236 years the nanosecond sum overflows int64 and wraps
		// negative, and a negative `expirationMs` is treated as already-expired by
		// `TimingWheel.add`, which fires the task immediately. Splitting both `now` and
		// `delay` into a whole-millisecond part and a sub-millisecond remainder keeps every
		// intermediate value small, so the result is exact across the full range of
		// `time.Duration` (up to ~292 years).
		expirationMs: expirationMs,
	}
	task.setBelongTo(te)
	return te
}

// ExpirationMs return the expiration milliseconds.
func (te *taskEntry) ExpirationMs() int64 { return te.expirationMs }

func (te *taskEntry) remove() {
	// If remove is called when another thread is moving the entry from a task entry list to another,
	// this may fail to remove the entry due to the change of value of list. Thus, we retry until the list becomes null.
	// In a rare case, this thread sees null and exits the loop, but the other thread insert the entry to another list later.
	for currentList := te.list.Load(); currentList != nil; currentList = te.list.Load() {
		currentList.Remove(te)
	}
}

func (te *taskEntry) cancelled() bool {
	return !te.task.isBelongTo(te)
}

// activated return true if the task entry is activated.
func (te *taskEntry) activated() bool {
	return te.list.Load() != nil
}
