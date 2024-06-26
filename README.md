# timer

Go implementation of Kafka's Hierarchical Timing Wheels.

[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/things-go/timer?tab=doc)
[![codecov](https://codecov.io/gh/things-go/timer/branch/main/graph/badge.svg)](https://codecov.io/gh/things-go/timer)
[![Tests](https://github.com/things-go/timer/actions/workflows/ci.yml/badge.svg)](https://github.com/things-go/timer/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/things-go/timer)](https://goreportcard.com/report/github.com/things-go/timer)
[![Licence](https://img.shields.io/github/license/things-go/timer)](https://raw.githubusercontent.com/things-go/timer/main/LICENSE)
[![Tag](https://img.shields.io/github/v/tag/things-go/timer)](https://github.com/things-go/timer/tags)

## Usage

### Installation

Use go get.

```bash
    go get github.com/things-go/timer
```

Then import the package into your own code.

```bash
    import "github.com/things-go/timer"
```

### Example

[embedmd]:# (examples/main.go go)
```go
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
```

## References

- [kafka timing wheels](https://github.com/apache/kafka/tree/trunk/server-common/src/main/java/org/apache/kafka/server/util/timer)

## License

This project is under MIT License. See the [LICENSE](LICENSE) file for the full license text.
