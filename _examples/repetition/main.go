package main

import (
	"fmt"
	"time"

	"github.com/thinkgos/timer/timed"
)

func main() {
	task := timed.NewTask(time.Second)
	task.WithJobFunc(func() {
		now := time.Now().String()
		timed.AddTask(task)
		fmt.Printf("%s: repetition executed,\n", now)
	})

	timed.AddTask(task)
	select {}
}
