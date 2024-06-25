package timer

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

// Task 是双向链表的一个元素.
type Task struct {
	// next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	prev, next *Task
	list       atomic.Pointer[Spoke] // 此元素所属的列表

	// follow The value stored with this element.
	delayMs      int64       // 延迟多少ms
	expirationMs int64       // 到期时间, 绝对时间, 单位: ms
	job          Job         // 任务
	hasCancelled atomic.Bool // 是否取消
}

// NewTask 创建一个空job任务条目
func NewTask(delayMs int64) *Task {
	return &Task{
		delayMs:      delayMs,
		expirationMs: delayMs + time.Now().UnixMilli(),
		job:          EmptyJob{},
	}
}

func NewTaskFunc(delayMs int64, f func()) *Task {
	return NewTask(delayMs).WithJobFunc(f)
}

func (t *Task) WithJobFunc(f func()) *Task {
	t.job = JobFunc(f)
	return t
}

func (t *Task) WithJob(j Job) *Task {
	t.job = j
	return t
}

func (t *Task) Run() {
	// hold recover
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintf(os.Stderr, "timer: Recovered from panic: %v", err)
		}
	}()
	t.job.Run()
}

func (t *Task) Cancel() {
	if t.list.Load() != nil {
		t.removeSelf()
		t.hasCancelled.Store(true)
	}
}

func (t *Task) cancelled() bool { return t.hasCancelled.Load() }

// nextTask 返回列表上的下一项, 如果没有返回nil
func (t *Task) nextTask() *Task {
	if p, list := t.next, t.list.Load(); list != nil && p != &list.root {
		return p
	}
	return nil
}

func (t *Task) removeSelf() {
	// If remove is called when another thread is moving the entry from a task entry list to another,
	// this may fail to remove the entry due to the change of value of list. Thus, we retry until the list becomes null.
	// In a rare case, this thread sees null and exits the loop, but the other thread insert the entry to another list later.
	for currentList := t.list.Load(); currentList != nil; currentList = t.list.Load() {
		currentList.Remove(t)
	}
}
