package timer

import (
	"time"

	"github.com/hashicorp/cronexpr"
)

var (
	_        Schedule = OneshotSchedule{}
	_        Schedule = PeriodicSchedule{}
	_        Schedule = CrontabSchedule{}
	emptyJob Schedule = Oneshot(JobFunc(func() {}))
)

// Job job interface
type Job interface {
	Run()
}

// JobFunc job function
type JobFunc func()

// Run implement job interface
func (f JobFunc) Run() { f() }

type Schedule interface {
	// NextDelay returns the delay duration.
	// NextDelay is invoked initially, and then each time the job is run.
	// If Delay returns less than 0 or equal to 0, the job will be canceled.
	NextDelay() time.Duration
	Job
}

// OneshotSchedule run a job once.
type OneshotSchedule struct {
	job Job
}

// Oneshot returns a ScheduleJob that runs a job once.
func Oneshot(job Job) Schedule {
	return OneshotSchedule{job: job}
}

func (j OneshotSchedule) Run() {
	j.job.Run()
}
func (j OneshotSchedule) NextDelay() time.Duration {
	return -1
}

// PeriodicSchedule represents a simple recurring duty cycle, e.g. "Every 5 minutes".
// It does not support jobs more frequent than once a second.
type PeriodicSchedule struct {
	job      Job
	duration time.Duration
}

// Periodic returns a periodic Schedule that activates once every duration.
// Delays of less than a second are not supported (will round up to 1 second).
// Any fields less than a Second are truncated.
func Periodic(duration time.Duration, job Job) Schedule {
	duration = max(duration, time.Second)
	return PeriodicSchedule{
		job:      job,
		duration: duration - time.Duration(duration.Nanoseconds())%time.Second,
	}
}

func (s PeriodicSchedule) Run() {
	s.job.Run()
}
func (s PeriodicSchedule) NextDelay() time.Duration {
	return s.duration
}

// CrontabSchedule specifies a duty cycle, based on a traditional crontab specification.
type CrontabSchedule struct {
	expr *cronexpr.Expression
	job  Job
}

func (s CrontabSchedule) Run() {
	s.job.Run()
}

func (s CrontabSchedule) NextDelay() time.Duration {
	now := time.Now()
	next := s.expr.Next(now)
	if next.IsZero() {
		return -1
	}
	return max(next.Sub(now), time.Second)
}

// Crontab returns a new crontab schedule representing the given spec.
// It returns a descriptive error if the spec is not valid.
// See <https://github.com/hashicorp/cronexpr#implementation> for documentation
// about what is a well-formed cron expression from this library's point of
// view.
func Crontab(spec string, job Job) (Schedule, error) {
	expr, err := cronexpr.Parse(spec)
	if err != nil {
		return nil, err
	}
	return CrontabSchedule{expr: expr, job: job}, nil
}
