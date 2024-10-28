package delayoverlapping

import (
	"context"
	"sync"
	"time"

	"github.com/flc1125/go-cron/v4"
)

type options struct {
	logger       cron.Logger
	reminderTime time.Duration
}

type Option func(*options)

func WithLogger(logger cron.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

func WithReminderTime(d time.Duration) Option {
	return func(o *options) {
		o.reminderTime = d
	}
}

func newOptions(opts ...Option) options {
	opt := options{
		logger:       cron.DefaultLogger,
		reminderTime: time.Minute,
	}
	for _, o := range opts {
		o(&opt)
	}
	return opt
}

// New returns a delay Overlapping middleware.
// if the job is running, delay the job.
// Based on the old version of DelayIfStillRunning
func New(opts ...Option) cron.Middleware {
	o := newOptions(opts...)
	return func(job cron.Job) cron.Job {
		var mu sync.Mutex
		return cron.JobFunc(func(ctx context.Context) error {
			mu.Lock()
			defer mu.Unlock()
			defer func(starting time.Time) {
				if dur := time.Since(starting); dur > o.reminderTime {
					o.logger.Info("delay", "duration", dur)
				}
			}(time.Now())

			return job.Run(ctx)
		})
	}
}
