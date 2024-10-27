package cron

import (
	"context"
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

// NoopMiddleware returns a Middleware that does nothing.
// It is useful for testing and for composing with other Middlewares.
func NoopMiddleware() Middleware {
	return func(j Job) Job {
		return j
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
