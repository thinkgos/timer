package main

import (
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/thinkgos/timer"

	_ "net/http/pprof"
)

type po struct {
	p *ants.Pool
}

func (s *po) Go(f func()) {
	s.p.Submit(f)
}

func newPo() *po {
	var defaultAntsPool, _ = ants.NewPool(ants.DefaultAntsPoolSize)
	return &po{p: defaultAntsPool}
}

var defaultPool = newPo()
var tim = timer.NewTimer(timer.WithGoPool(defaultPool))

func init() {
	tim.Start()
}

func main() {
	go func() {
		sum := &atomic.Int64{}
		t := time.NewTicker(time.Second)
		for {
			<-t.C
			added := 0
			ranv := rand.IntN(10)
			max := int(rand.Uint32N(math.MaxUint16))
			for i := 100; i < max; i += 100 {
				added++
				ii := i + ranv

				defaultPool.Go(func() {
					sum.Add(1)
					delayms := int64(ii) * 20
					task := timer.NewTask(time.Duration(delayms) * time.Millisecond).WithJob(&job{
						sum:          sum,
						expirationMs: time.Now().UnixMilli() + delayms,
					})
					tim.AddTask(task)

					// for test race
					// if ii%0x03 == 0x00 {
					// 	defaultPool.Go(func() {
					// 		task.Cancel()
					// 	})
					// }
				})
			}
			log.Printf("task: %v - %v added: %d", tim.TaskCounter(), sum.Load(), added)
		}
	}()

	addr := ":9990"
	log.Printf("http stated '%v'\n", addr)
	log.Println(http.ListenAndServe(addr, nil))
}

type job struct {
	sum          *atomic.Int64
	expirationMs int64
}

func (j *job) Run() {
	j.sum.Add(-1)
	now := time.Now().UnixMilli()
	if diff := now - j.expirationMs; diff > 1 {
		log.Printf("this task no equal, diff: %d %d %d\n", now, j.expirationMs, diff)
	}
}
