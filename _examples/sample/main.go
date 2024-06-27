package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/thinkgos/timer"
)

func main() {
	t := timer.NewTimer()
	t.Start()
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		index := i
		t.AfterFunc(time.Duration(i)*time.Millisecond, func() {
			fmt.Printf("%d: timer task %d is executed, remain task: %d\n", time.Now().UnixMilli(), index, t.TaskCounter())
			wg.Done()
		})
	}
	wg.Wait()
	t.Stop()
}
