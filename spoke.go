package timer

import (
	"sync"
	"sync/atomic"
	"time"
)

// Spoke a spoke of the wheel.
type Spoke struct {
	root        taskEntry     // sentinel list element, only &root, root.prev, and root.next are used
	expiration  atomic.Int64  // the expiration time
	mu          sync.Mutex    // protects all list's action.
	taskCounter *atomic.Int64 // same as Timer.taskCounter
}

func NewSpoke(taskCounter *atomic.Int64) *Spoke {
	sp := &Spoke{
		taskCounter: taskCounter,
	}
	sp.expiration.Store(-1)
	sp.root.next = &sp.root
	sp.root.prev = &sp.root
	return sp
}

// Add the timer task to this list
func (sp *Spoke) Add(te *taskEntry) {
	for done := false; !done; {
		// Remove the timer task if it is already in any other list
		// We do this outside of the sync block below to avoid deadlocking.
		// We may retry until task.list becomes null.
		te.remove()
		if te.list.Load() == nil { // fast check.
			sp.mu.Lock()
			if te.list.Load() == nil { // double check but slow.
				sp.pushBack(te)
				done = true
			}
			sp.mu.Unlock()
		}
	}
}

// Remove the specified timer task from this list
func (sp *Spoke) Remove(te *taskEntry) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if te.list.Load() == sp {
		sp.remove(te)
	}
}

func (sp *Spoke) pushBack(te *taskEntry) {
	at := sp.root.prev

	te.prev = at
	te.next = at.next
	te.prev.next = te
	te.next.prev = te
	te.list.Store(sp)
	sp.taskCounter.Add(1)
}

func (sp *Spoke) remove(te *taskEntry) {
	te.prev.next = te.next
	te.next.prev = te.prev
	te.next = nil // avoid memory leaks
	te.prev = nil // avoid memory leaks
	te.list.Store(nil)
	sp.taskCounter.Add(-1)
}

// Flush all task entries and apply the supplied function to each of them.
//
// NOTE: `f` is applied OUTSIDE of the spoke's lock, on purpose.
// `f` is `Timer.addTaskEntry`, which acquires `Task.rw` (@taskEntry.cancelled),
// while `Task.Cancel` acquires `Task.rw` before this spoke's lock (@taskEntry.remove).
// Applying `f` under `sp.mu` inverts that order and deadlocks.
// It is the same reason `Spoke.Add` calls `taskEntry.remove` outside of the lock.
//
// NOTE: the detached chain itself is the work list, so a flush allocates nothing.
// That is also why the walk does not reuse `Spoke.remove`: it clears `prev`/`next`,
// which are the only handle on the entries still to be visited, and it writes
// `e.prev.next`, which for the chain's head is `sp.root.next`, a write to the
// sentinel outside the lock. The walk touches only `e`'s own fields, and compares
// against `&sp.root` without ever reading through it.
//
// NOTE: clearing an entry's links outside the lock is safe only because nothing can
// re-insert a chain entry behind the walk's back. `Spoke.Add` is reachable only from
// `Timer.addTaskEntry`, whose callers are this flush itself (same goroutine, and only
// for the entry it is currently visiting) and `Timer.AddTask`, which always passes a
// fresh `newTaskEntry` and is locked out of `Timer.rw` while the flush runs. Adding a
// caller that re-inserts an entry the flush still holds would leave that entry in the
// ring with nil links, and the next `Spoke.remove` would dereference them.
func (sp *Spoke) Flush(f func(*taskEntry)) {
	sp.mu.Lock()
	// Splice the whole chain off the sentinel in one step, then reset the sentinel so
	// the spoke reads as empty. The chain keeps its links, and its ends keep pointing
	// at `&sp.root`, which the walk below only ever compares against, never follows.
	first := sp.root.next
	sp.root.next, sp.root.prev = &sp.root, &sp.root
	// Detach every entry logically. Once `taskEntry.list` is nil the entry no longer
	// belongs to this spoke, so a concurrent `Task.Cancel` will not contend on `sp.mu`
	// and, from here on, nothing but the walk below touches the chain's links.
	//
	// The successor is read BEFORE the entry is detached, on purpose: while `e.list`
	// still reports this spoke, the only way to reach `e.next` is `Spoke.remove`, which
	// needs `sp.mu`. Reading it after the store leaves it unprotected, and a `Spoke.Add`
	// of `e` onto another spoke, which takes only that spoke's lock, could redirect this
	// loop into a live ring and detach entries that are not ours.
	detached := int64(0)
	for e := first; e != &sp.root; {
		next := e.next
		e.list.Store(nil)
		detached++
		e = next
	}
	// One update instead of one per entry: the counter is shared by every spoke of every
	// wheel (@Timer.taskCounter), so the whole flush should touch it once, and not at
	// all when the spoke came up empty, which it does whenever every task on it was
	// cancelled before it expired.
	if detached != 0 {
		sp.taskCounter.Add(-detached)
	}
	// Reset the expiration before releasing the lock, so that a re-inserted task
	// landing back on this spoke observes a changed expiration and gets enqueued.
	sp.SetExpiration(-1)
	sp.mu.Unlock()

	for e := first; e != &sp.root; {
		next := e.next            // `f` may re-insert `e`, which overwrites its links.
		e.next, e.prev = nil, nil // avoid memory leaks
		f(e)
		e = next
	}
}

// SetExpiration set the spoke's expiration time
// Returns true if the expiration time changes.
func (sp *Spoke) SetExpiration(expirationMs int64) bool {
	return sp.expiration.Swap(expirationMs) != expirationMs
}

// GetExpiration the spoke's expiration time
func (sp *Spoke) GetExpiration() int64 { return sp.expiration.Load() }

// Delay implements delayqueue.Delayed.
func (sp *Spoke) Delay() int64 {
	delay := sp.GetExpiration() - time.Now().UnixMilli()
	if delay < 0 {
		return 0
	}
	return delay
}

// CompareSpoke compares two Spoke instances based on their expiration time.
func CompareSpoke(sp1, sp2 *Spoke) int {
	v1, v2 := sp1.GetExpiration(), sp2.GetExpiration()
	if v1 < v2 {
		return -1
	}
	if v1 > v2 {
		return 1
	}
	return 0
}
