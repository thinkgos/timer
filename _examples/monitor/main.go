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

var tim = timer.NewTimer(timer.WithGoPool(newPo()))

func init() {
	tim.Start()
}

func main() {
	go func() {
		sum := &atomic.Int64{}
		t := time.NewTicker(time.Second)
		for {
			<-t.C
			plus := rand.Int() & 0x08
			now := time.Now().UnixMilli()
			for i := 1; i < int(rand.Uint32N(math.MaxUint16)); i = i + plus {
				sum.Add(1)
				delayms := int64(i) * 100
				j := &job{
					sum:          sum,
					expirationMs: now + delayms,
				}
				tim.AddTask(timer.NewTask(delayms).WithJob(j))
			}
			log.Printf("task: %v - %v", tim.TaskCounter(), sum.Load())
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
