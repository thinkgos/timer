# timer

Go implementation of Kafka's Hierarchical Timing Wheels.

[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/thinkgos/timer?tab=doc)
[![codecov](https://codecov.io/gh/thinkgos/timer/branch/main/graph/badge.svg)](https://codecov.io/gh/thinkgos/timer)
[![Tests](https://github.com/thinkgos/timer/actions/workflows/ci.yml/badge.svg)](https://github.com/thinkgos/timer/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/thinkgos/timer)](https://goreportcard.com/report/github.com/thinkgos/timer)
[![License](https://img.shields.io/github/license/thinkgos/timer)](https://raw.githubusercontent.com/thinkgos/timer/main/LICENSE)
[![Tag](https://img.shields.io/github/v/tag/thinkgos/timer)](https://github.com/thinkgos/timer/tags)

## Feature

- Unlimited hierarchical wheel.
- `insert`, `delete`, `scan` task almost O(1).
- Different from the time wheel of Linux, it has no maximum time limit.
- It is not advancing per **TickMs**, it uses `DelayQueue` to directly take out the most recently expired `Spoke`, and then advances to the expiration time of the `Spoke` in one step, preventing empty advances.
- built-in a global `timer` instance, that tick is 1ms. wheel size is 128, use [ants](https://github.com/panjf2000/ants) goroutine pool.

## Usage

### Installation

Use go get.

```bash
    go get github.com/thinkgos/timer
```

Then import the package into your own code.

```bash
    import "github.com/thinkgos/timer"
```

### Example

- [monitor](./_examples/monitor/main.go)
- [repetition](./_examples/repetition/main.go)
- [sample](./_examples/sample/main.go)

#### monitor

[embedmd]:# (_examples/monitor/main.go go)
```go
package main

import (
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"sync/atomic"
	"time"

	_ "net/http/pprof"

	"github.com/thinkgos/timer"
)

// almost 1,000,000 task
func main() {
	go func() {
		sum := &atomic.Int64{}
		t := time.NewTicker(time.Second)
		for {
			<-t.C
			added := 0
			ranv := rand.IntN(10)
			max := int(rand.Uint32N(math.MaxUint16 << 2))
			for i := 100; i < max; i += 200 {
				added++
				ii := i + ranv

				timer.Go(func() {
					sum.Add(1)
					delayms := int64(ii) * 20
					task := timer.NewTask(time.Duration(delayms) * time.Millisecond).WithJob(&job{
						sum:          sum,
						expirationMs: time.Now().UnixMilli() + delayms,
					})
					timer.AddTask(task)

					// for test race
					// if ii%0x03 == 0x00 {
					// 	timer.Go(func() {
					// 		task.Cancel()
					// 	})
					// }
				})
			}
			log.Printf("task: %v - %v added: %d", timer.TaskCounter(), sum.Load(), added)
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
```

#### repetition

[embedmd]:# (_examples/repetition/main.go go)
```go
package main

import (
	"fmt"
	"time"

	"github.com/thinkgos/timer"
)

// one or two second delay repetition example
func main() {
	job := NewRepetitionJob()
	_ = timer.AddDerefTask(job)
	select {}
}

type RepetitionJob struct {
	task *timer.Task
	i    int
}

var _ timer.TaskContainer = (*RepetitionJob)(nil)

func NewRepetitionJob() *RepetitionJob {
	j := &RepetitionJob{
		task: timer.NewTask(time.Second),
		i:    1,
	}
	j.task.WithJob(j)
	return j
}

func (j *RepetitionJob) Run() {
	now := time.Now().String()
	j.i++
	_ = timer.AddTask(j.task.SetDelay(time.Second * time.Duration((j.i%2 + 1))))
	fmt.Printf("%s: repetition executed,\n", now)
}

func (j *RepetitionJob) DerefTask() *timer.Task { return j.task }
```

#### sample

[embedmd]:# (_examples/sample/main.go go)
```go
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
```



## How it works

- [How it works](./how_it_works.md)

## References

- [kafka timing wheels](https://github.com/apache/kafka/tree/trunk/server-common/src/main/java/org/apache/kafka/server/util/timer)

## License

This project is under MIT License. See the [LICENSE](LICENSE) file for the full license text.
