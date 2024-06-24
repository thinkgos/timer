package main

import (
	"log"
	"sync"
	"time"

	"github.com/things-go/timer"
)

func main() {
	t := timer.NewTimer()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		index := i
		t.AfterFunc(time.Duration(i)*time.Second, func() {
			log.Printf("timer task %d is executed\n", index)
			wg.Done()
		})
	}
	t.Start()
	wg.Wait()
	t.Stop()
}
