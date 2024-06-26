package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/things-go/timer"
)

func main() {
	t := timer.NewTimer()
	t.Start()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		index := i
		t.AfterFunc(time.Duration(i)*time.Second+time.Duration(i)*10*time.Millisecond, func() {
			fmt.Printf("%s: timer task %d is executed, remain task: %d\n", time.Now().String(), index, t.TaskCounter())
			wg.Done()
		})
	}
	wg.Wait()
	t.Stop()
}
