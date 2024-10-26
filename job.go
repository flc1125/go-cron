package cron

import "context"

// Job is an interface for submitted cron jobs.
type Job interface {
	Run(ctx context.Context) error
}

// JobFunc is a wrapper that turns a func(context.Context) into a cron.Job
type JobFunc func(ctx context.Context) error

func (fn JobFunc) Run(ctx context.Context) error {
	return fn(ctx)
}

// NoopJob is a job that does nothing.
// it is useful for testing and examples.
type NoopJob struct{}

func (NoopJob) Run(context.Context) error {
	return nil
}
