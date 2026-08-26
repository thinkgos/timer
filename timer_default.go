package timer

import (
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

var pool GoPool = wrapperAnts{}

// defaultTimer is constructed but not started at package load. Construction only
// allocates the wheel; the goroutine that drives it is spawned by the first call
// to `DefaultTimer` (or a mutating helper), so merely importing the package has
// no background side effect.
var defaultTimer = NewTimer(WithGoPool(pool))

// startOnce gates the automatic start on first use.
//
// NOTE: it fires exactly once, for the FIRST use only. Once consumed it never fires
// again, so after an explicit `Stop()` the auto-start helpers (`DefaultTimer`,
// `AfterFunc`, `AddTask`, `AddDerefTask`) can no longer bring the timer back up:
// they keep returning `ErrClosed` until `Start()` is called explicitly. In other
// words, "start on first use" holds exactly once; restarting after a stop is always
// the caller's job. see `startDefault`, `Stop`.
var startOnce sync.Once

// startDefault ensures the default timer's goroutine is running. It is idempotent
// and cheap on the hot path: `sync.Once` collapses to a single atomic load after
// the first call, and `Timer.Start` is itself idempotent. Explicit `Start`/`Stop`
// bypass it, because the timer must be restartable after a `Stop`.
//
// Because `sync.Once` is consumed by this first automatic start, the guarantee only
// covers the transition from never-started to started. Once the timer has been
// stopped, this function is a no-op and `defaultTimer` stays closed; use the package
// level `Start()` to restart it.
func startDefault() {
	startOnce.Do(func() {
		defaultTimer.Start()
	})
}

// DefaultTimer return the default timer, started on first call.
func DefaultTimer() *Timer {
	startDefault()
	return defaultTimer
}

// TickMs return Basic time tick milliseconds.
func TickMs() int64 { return defaultTimer.TickMs() }

// WheelSize return the wheel size.
func WheelSize() int { return defaultTimer.WheelSize() }

// TaskCounter return the total number of tasks.
func TaskCounter() int64 { return defaultTimer.TaskCounter() }

// AfterFunc adds a function to the timer, starting it on first use.
//
// NOTE: the automatic start happens only on the FIRST use (see `startDefault`).
// After an explicit `Stop()`, this returns `ErrClosed` instead of restarting the
// timer; call `Start()` first to use it again.
func AfterFunc(d time.Duration, f func()) (*Task, error) {
	return DefaultTimer().AfterFunc(d, f)
}

// AddTask adds a task to the timer, starting it on first use.
//
// NOTE: the automatic start happens only on the FIRST use (see `startDefault`).
// After an explicit `Stop()`, this returns `ErrClosed` instead of restarting the
// timer; call `Start()` first to use it again.
func AddTask(task *Task) error {
	return DefaultTimer().AddTask(task)
}

// AddDerefTask adds a task from DerefTask to the timer, starting it on first use.
//
// NOTE: the automatic start happens only on the FIRST use (see `startDefault`).
// After an explicit `Stop()`, this returns `ErrClosed` instead of restarting the
// timer; call `Start()` first to use it again.
func AddDerefTask(task DerefTask) error {
	return DefaultTimer().AddDerefTask(task)
}

// Started have started or not.
func Started() bool { return defaultTimer.Started() }

// Start the timer. It is idempotent, and is the only way to bring the default
// timer back up after a `Stop` (the auto-start helpers do not restart it, see
// `startDefault`).
func Start() { defaultTimer.Start() }

// Stop the timer.
//
// NOTE: after `Stop`, the auto-start helpers (`AfterFunc`, `AddTask`,
// `AddDerefTask`) return `ErrClosed` instead of restarting the timer, because the
// one-shot automatic start has already been consumed. Restart with `Start()`.
func Stop() { defaultTimer.Stop() }

type wrapperAnts struct{}

func (wrapperAnts) Go(f func()) {
	Go(f)
}

// Go run a function in `ants` goroutine pool, if submit failed, fallback to use goroutine.
func Go(f func()) {
	if err := ants.Submit(f); err != nil {
		go f()
	}
}
