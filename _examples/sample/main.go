package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/thinkgos/timer"
)

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		index := i
		_, _ = timer.AfterFunc(time.Duration(i)*100*time.Millisecond, func() {
			fmt.Printf("%s: timer task %d is executed, remain task: %d\n", time.Now().String(), index, timer.TaskCounter())
			wg.Done()
		})
	}
	wg.Wait()
}
