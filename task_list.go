package timer

import (
	"sync"
	stdAtomic "sync/atomic"

	"go.uber.org/atomic"
)

type TaskList struct {
	*list
	expiration int64
	sync.Mutex
}

func NewTaskList(counter *atomic.Int64) *TaskList {
	return &TaskList{
		list: newList(counter),
	}
}

func (t *TaskList) Add(e *TaskEntry) {
	t.Lock()
	defer t.Unlock()
	t.PushBack(e)
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

// Remove all task entries and apply the supplied function to each of them
func (l *TaskList) Flush(f func(entry *TaskEntry)) {
	l.Lock()
	defer l.Unlock()
	for e := l.Front(); e != nil; e = e.Next() {
		l.remove(e)
		f(e)
	}
	l.SetExpiration(-1)
}

func (l *TaskList) DelayMs() int64 {
	delay := l.GetExpiration() - NowMs()
	if delay < 0 {
		return 0
	}
	return delay
}

func CompareTaskList(v1, v2 interface{}) int {
	d1, d2 := v1.(*TaskList).GetExpiration(), v2.(*TaskList).GetExpiration()
	if d1 < d2 {
		return -1
	}
	if d1 > d2 {
		return 1
	}
	return 0
}
