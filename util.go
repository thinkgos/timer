package timer

// Job job interface
type Job interface {
	Run()
}

type GoPool interface {
	Go(f func())
}

// JobFunc job function
type JobFunc func()

// Run implement job interface
func (f JobFunc) Run() { f() }

type EmptyJob struct{}

func (EmptyJob) Run() {}

type InternalGoPool struct{}

func (InternalGoPool) Go(f func()) {
	go f()
}

func IsPowOf2(x int) bool {
	return (x & (x - 1)) == 0
}

func NextPowOf2(x int) int {
	if IsPowOf2(x) {
		return x
	}
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	return x + 1
}
