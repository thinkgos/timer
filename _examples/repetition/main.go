package main

import (
	"fmt"
	"time"

	"github.com/thinkgos/timer"
)

// one or two second delay repetition example
func main() {
	s := NewRepetitionSchedule()
	task := timer.NewScheduleTask(s.NextDelay(), s)
	_ = timer.AddTask(task)
	select {}
}

type RepetitionJob struct {
	i int
}

func NewRepetitionSchedule() timer.Schedule {
	return &RepetitionJob{
		i: 0,
	}
}

func (j *RepetitionJob) NextDelay() time.Duration {
	return time.Second * time.Duration((j.i%2 + 1))
}

func (j *RepetitionJob) Run() {
	now := time.Now().String()
	j.i++
	fmt.Printf("%s: repetition executed,\n", now)
}
