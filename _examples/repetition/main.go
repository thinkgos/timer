package main

import (
	"fmt"
	"time"

	"github.com/thinkgos/timer"
)

// one or two second delay repetition example
func main() {
	job := &RepetitionJob{
		task: timer.NewTask(time.Second),
		i:    1,
	}
	job.task.WithJob(job)
	_ = timer.AddTask(job.task)
	select {}
}

type RepetitionJob struct {
	task *timer.Task
	i    int
}

func (j *RepetitionJob) Run() {
	now := time.Now().String()
	j.i++
	_ = timer.AddTask(j.task.SetDelay(time.Second * time.Duration((j.i%2 + 1))))
	fmt.Printf("%s: repetition executed,\n", now)
}
