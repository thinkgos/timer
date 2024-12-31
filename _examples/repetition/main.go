package main

import (
	"fmt"
	"time"

	"github.com/thinkgos/timer"
)

// one or two second delay repetition example
func main() {
	i := 0
	task := timer.NewTask(time.Second)
	task.WithJobFunc(func() {
		now := time.Now().String()
		i++
		timer.AddTask(task.SetDelay(time.Second * time.Duration((i%2 + 1))))
		fmt.Printf("%s: repetition executed,\n", now)
	})
	timer.AddTask(task)
	select {}
}
