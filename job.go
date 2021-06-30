package timer

// Job job interface
type Job interface {
	Run()
}

// JobFunc job function
type JobFunc func()

// Run implement job interface
func (f JobFunc) Run() { f() }

// hold recover
func wrapRunJob(job Job) {
	defer func() {
		_ = recover()
	}()
	job.Run()
}
