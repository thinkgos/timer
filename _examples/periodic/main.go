package main

import (
	"log"
	"time"

	"github.com/thinkgos/timer"
)

func main() {
	// t, err := schedule.NewCrontabTask("@hourly", timer.JobFunc(func() {

	// 	log.Println("hello world!!")
	// }))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	t := timer.NewPeriodicTask(time.Second*2, timer.JobFunc(func() {
		log.Println("hello world!!")
	}))
	timer.AddDerefTask(t)
	select {}
}
