package cron

import "slices"

// Middleware is a function that wraps a Job to provide additional functionality.
type Middleware func(Job) Job

// Chain is a helper function to compose Middlewares. It returns a Middleware that
// applies the Middlewares in order.
//
//	Chain(m1, m2, m3) => m1(m2(m3(job)))
func Chain(m ...Middleware) Middleware {
	return func(next Job) Job {
		for _, v := range slices.Backward(m) {
			next = v(next)
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
