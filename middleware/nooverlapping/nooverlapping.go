package nooverlapping

import (
	"context"

	"github.com/flc1125/go-cron/v4"
)

type options struct {
	logger cron.Logger
}

type Option func(*options)

func WithLogger(logger cron.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

func newOptions(opts ...Option) options {
	opt := options{
		logger: cron.DefaultLogger,
	}
	for _, o := range opts {
		o(&opt)
	}
	return opt
}

// New returns a without Overlapping middleware.
// if the job is running, skip the job.
// Based on the old version of SkipIfStillRunning
func New(opts ...Option) cron.Middleware {
	o := newOptions(opts...)
	return func(job cron.Job) cron.Job {
		ch := make(chan struct{}, 1)
		ch <- struct{}{}
		return cron.JobFunc(func(ctx context.Context) error {
			select {
			case v := <-ch:
				defer func() { ch <- v }()
				return job.Run(ctx)
			default:
				o.logger.Info("job is still running, skip")
				return nil
			}
		})
	}
}
