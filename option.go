package cron

import (
	"context"
	"time"
)

// Option represents a modification to the default behavior of a Cron.
type Option func(*Cron)

// WithContext overrides the context used by the Cron.
func WithContext(ctx context.Context) Option {
	return func(c *Cron) {
		c.ctx = ctx
	}
}

// WithLocation overrides the timezone of the cron instance.
func WithLocation(loc *time.Location) Option {
	return func(c *Cron) {
		c.location = loc
	}
}

// WithSeconds overrides the parser used for interpreting job schedules to
// include a seconds field as the first one.
func WithSeconds() Option {
	return WithParser(NewParser(
		Second | Minute | Hour | Dom | Month | Dow | Descriptor,
	))
}

// WithParser overrides the parser used for interpreting job schedules.
func WithParser(p ScheduleParser) Option {
	return func(c *Cron) {
		c.parser = p
	}
}

// WithMiddleware specifies Middleware to apply to all jobs added to this cron.
func WithMiddleware(middlewares ...Middleware) Option {
	return func(c *Cron) {
		c.middlewares = middlewares
	}
}

// WithLogger uses the provided logger.
func WithLogger(logger Logger) Option {
	return func(c *Cron) {
		c.logger = logger
	}
}
