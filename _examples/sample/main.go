package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/thinkgos/timer/timed"
)

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		index := i
		_, _ = timed.AfterFunc(time.Duration(i)*100*time.Millisecond, func() {
			fmt.Printf("%s: timer task %d is executed, remain task: %d\n", time.Now().String(), index, timed.TaskCounter())
			wg.Done()
		})
	}
	wg.Wait()
}
