package main

import (
	"fmt"
	"time"

	"github.com/thinkgos/timer/timed"
)

// one or two second delay repetition example
func main() {
	i := 0
	task := timed.NewTask(time.Second)
	task.WithJobFunc(func() {
		now := time.Now().String()
		i++
		timed.AddTask(task.SetDelay(time.Second * time.Duration((i%2 + 1))))
		fmt.Printf("%s: repetition executed,\n", now)
	})
	timed.AddTask(task)
	select {}
}
