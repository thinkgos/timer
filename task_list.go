package timer

import (
	"sync"
	stdAtomic "sync/atomic"

	"go.uber.org/atomic"
)

type TaskList struct {
	root       *list
	expiration int64
	sync.Mutex
}

func NewTaskList(counter *atomic.Int64) *TaskList {
	return &TaskList{
		root: newList(counter),
	}
}

func (t *TaskList) Add(e *TaskEntry) {
	t.Lock()
	defer t.Unlock()
	e.removeSelf()
	t.root.PushElementBack(e)
}

// Set the bucket's expiration time
// Returns true if the expiration time is changed
func (t *TaskList) SetExpiration(expirationMs int64) bool {
	return stdAtomic.SwapInt64(&t.expiration, expirationMs) != expirationMs
}

// Get the bucket's expiration time
func (l *TaskList) GetExpiration() int64 {
	return stdAtomic.LoadInt64(&l.expiration)
}

func (l *TaskList) DelayMs() int64 {
	delay := l.GetExpiration() - NowMs()
	if delay < 0 {
		return 0
	}
	return delay
}

func CompareTaskList(v1, v2 interface{}) int {
	l1, l2 := v1.(*TaskList), v2.(*TaskList)
	if l1.GetExpiration() < l2.GetExpiration() {
		return -1
	}
	if l1.GetExpiration() > l2.GetExpiration() {
		return 1
	}
	return 0
}
