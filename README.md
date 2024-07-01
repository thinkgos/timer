# timer

Go implementation of Kafka's Hierarchical Timing Wheels, Support millions of tasks. `timed` is a global `timer` instance, that tick is 1ms. wheel size is 1024, use [ants](https://github.com/panjf2000/ants) goroutine pool.

[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/thinkgos/timer?tab=doc)
[![codecov](https://codecov.io/gh/thinkgos/timer/branch/main/graph/badge.svg)](https://codecov.io/gh/thinkgos/timer)
[![Tests](https://github.com/thinkgos/timer/actions/workflows/ci.yml/badge.svg)](https://github.com/thinkgos/timer/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/thinkgos/timer)](https://goreportcard.com/report/github.com/thinkgos/timer)
[![Licence](https://img.shields.io/github/license/thinkgos/timer)](https://raw.githubusercontent.com/thinkgos/timer/main/LICENSE)
[![Tag](https://img.shields.io/github/v/tag/thinkgos/timer)](https://github.com/thinkgos/timer/tags)

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

[embedmd]:# (_examples/sample/main.go go)
```go
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
```

## References

- [kafka timing wheels](https://github.com/apache/kafka/tree/trunk/server-common/src/main/java/org/apache/kafka/server/util/timer)

## License

This project is under MIT License. See the [LICENSE](LICENSE) file for the full license text.
