package timer

import (
	"time"
)

// Job job interface
type Job interface {
	Run()
}

// JobFunc job function
type JobFunc func()

// Run implement job interface
func (f JobFunc) Run() { f() }

func wrapJob(job Job) {
	defer func() {
		_ = recover()
	}()
	job.Run()
}

type emptyJob struct{}

func (emptyJob) Run() {}

func NowMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
