package main

import (
	"fmt"
	"time"

	"github.com/thinkgos/timer"
)

// one or two second delay repetition example
func main() {
	job := NewRepetitionJob()
	_ = timer.AddDerefTask(job)
	select {}
}

type RepetitionJob struct {
	task *timer.Task
	i    int
}

func NewRepetitionJob() *RepetitionJob {
	j := &RepetitionJob{
		task: timer.NewTask(time.Second),
		i:    1,
	}
	j.task.WithJob(j)
	return j
}

func (j *RepetitionJob) Run() {
	now := time.Now().String()
	j.i++
	_ = timer.AddTask(j.task.SetDelay(time.Second * time.Duration((j.i%2 + 1))))
	fmt.Printf("%s: repetition executed,\n", now)
}

func (j *RepetitionJob) DerefTask() *timer.Task { return j.task }
