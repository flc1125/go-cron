package cron

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// Middleware is a function that wraps a Job to provide additional functionality.
type Middleware func(Job) Job

// Chain is a helper function to compose Middlewares. It returns a Middleware that
// applies the Middlewares in order.
//
//	Chain(m1, m2, m3) => m1(m2(m3(job)))
func Chain(m ...Middleware) Middleware {
	return func(next Job) Job {
		for i := len(m) - 1; i >= 0; i-- {
			next = m[i](next)
		}
		return next
	}
}

// Recover panics in wrapped jobs and log them with the provided logger.
// Deprecated: recovery.New()
func Recover(logger Logger) Middleware {
	return func(j Job) Job {
		return JobFunc(func(ctx context.Context) error {
			defer func() {
				if r := recover(); r != nil {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					logger.Error(err, "panic", "stack", "...\n"+string(buf))
				}
			}()
			return j.Run(ctx)
		})
	}
}

// DelayIfStillRunning serializes jobs, delaying subsequent runs until the
// previous one is complete. Jobs running after a delay of more than a minute
// have the delay logged at Info.
func DelayIfStillRunning(logger Logger) Middleware {
	return func(j Job) Job {
		var mu sync.Mutex
		return JobFunc(func(ctx context.Context) error {
			start := time.Now()
			mu.Lock()
			defer mu.Unlock()
			if dur := time.Since(start); dur > time.Minute {
				logger.Info("delay", "duration", dur)
			}
			return j.Run(ctx)
		})
	}
}

// SkipIfStillRunning skips an invocation of the Job if a previous invocation is
// still running. It logs skips to the given logger at Info level.
func SkipIfStillRunning(logger Logger) Middleware {
	return func(j Job) Job {
		ch := make(chan struct{}, 1)
		ch <- struct{}{}
		return JobFunc(func(ctx context.Context) error {
			select {
			case v := <-ch:
				defer func() { ch <- v }()
				return j.Run(ctx)
			default:
				logger.Info("skip")
				return nil
			}
		})
	}
}
