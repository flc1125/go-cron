package cron

import "context"

// Job is an interface for submitted cron jobs.
type Job interface {
	Run(ctx context.Context)
}

// JobFunc is a wrapper that turns a func(context.Context) into a cron.Job
type JobFunc func(ctx context.Context)

func (j JobFunc) Run(ctx context.Context) { j(ctx) }
