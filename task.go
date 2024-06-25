package timer

import (
	"fmt"
	"os"
	"sync/atomic"
)

// TaskEntry 是双向链表的一个元素.
type TaskEntry struct {
	// next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	prev, next *TaskEntry
	list       *Spoke // 此元素所属的列表

	// follow The value stored with this element.
	delayMs      int64       // 延迟多少ms
	expirationMs int64       // 到期时间, 绝对时间, 单位: ms
	job          Job         // 任务
	cancelled    atomic.Bool // 是否取消
}

// nextEntry 返回列表上的下一项, 如果没有返回nil
func (e *TaskEntry) nextEntry() *TaskEntry {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// TODO: 优化
func (e *TaskEntry) removeSelf() {
	// If remove is called when another thread is moving the entry from a task entry list to another,
	// this may fail to remove the entry due to the change of value of list. Thus, we retry until the list becomes null.
	// In a rare case, this thread sees null and exits the loop, but the other thread insert the entry to another list later.
	for currentList := e.list; currentList != nil; currentList = e.list {
		currentList.Remove(e)
	}
}

// NewTaskEntry 创建一个空job任务条目
func NewTaskEntry(delayMs int64) *TaskEntry {
	return &TaskEntry{
		delayMs:      delayMs,
		expirationMs: delayMs + NowMs(),
		job:          EmptyJob{},
	}
}

func (t *TaskEntry) WithJobFunc(f func()) *TaskEntry {
	t.job = JobFunc(f)
	return t
}

func (t *TaskEntry) WithJob(j Job) *TaskEntry {
	t.job = j
	return t
}

func NewTaskEntryFunc(delayMs int64, f func()) *TaskEntry {
	return &TaskEntry{
		delayMs:      delayMs,
		expirationMs: delayMs + NowMs(),
		job:          JobFunc(f),
	}
}

func (t *TaskEntry) isCancelled() bool { return t.cancelled.Load() }

func (t *TaskEntry) Run() {
	// hold recover
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintf(os.Stderr, "timer: Recovered from panic: %v", err)
		}
	}()
	t.job.Run()
}

// TODO: 优化
func (t *TaskEntry) Cancel() {
	if t.list != nil {
		t.removeSelf()
		t.cancelled.Store(true)
	}
}
