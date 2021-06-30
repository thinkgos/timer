package goroutine

import (
	"time"

	"github.com/panjf2000/ants/v2"
)

const (
	// DefaultAntsPoolSize sets up the capacity of worker pool, 256 * 1024.
	DefaultAntsPoolSize = 1 << 18
	// ExpiryDuration is the interval time to clean up those expired workers.
	ExpiryDuration = 10 * time.Second
)

// Default instantiates a non-blocking *WorkerPool with the capacity of DefaultAntsPoolSize.
var Pool *ants.Pool

func init() {
	Pool, _ = ants.NewPool(DefaultAntsPoolSize, ants.WithOptions(ants.Options{ExpiryDuration: ExpiryDuration, Nonblocking: true}))
}

func Go(f func()) {
	err := Pool.Submit(f)
	if err != nil {
		go func() {
			defer func() {
				_ = recover()
			}()
			f()
		}()
	}
}
