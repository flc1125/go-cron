package cron

import (
	"context"
	"time"
)

// EntryID identifies an entry within a Cron instance
type EntryID int

// Entry consists of a schedule and the func to execute on that schedule.
type Entry struct {
	// id is the cron-assigned id of this entry, which may be used to look up a
	// snapshot or remove it.
	id EntryID

	// schedule on which this job should be run.
	schedule Schedule

	// next time the job will run, or the zero time if Cron has not been
	// started or this entry's schedule is unsatisfiable
	next time.Time

	// prev is the last time this job was run, or the zero time if never.
	prev time.Time

	// wrappedJob is the thing to run when the schedule is activated.
	wrappedJob Job

	// job is the thing that was submitted to cron.
	// It is kept around so that user code that needs to get at the job later,
	// e.g. via Entries() can do so.
	job Job

	// middlewares are the list of middlewares to apply to the job.
	middlewares []Middleware
}

type EntryOption func(*Entry)

func WithEntryMiddlewares(middlewares ...Middleware) EntryOption {
	return func(e *Entry) {
		e.middlewares = middlewares
	}
}

// newEntry creates a new entry with the given schedule and job.
func newEntry(id EntryID, schedule Schedule, job Job, opts ...EntryOption) *Entry {
	entry := &Entry{
		id:       id,
		schedule: schedule,
		job:      job,
	}
	for _, opt := range opts {
		opt(entry)
	}

	// Wrap the job with the entry context.
	middlewares := append([]Middleware{
		func(job Job) Job {
			return JobFunc(func(ctx context.Context) error {
				return job.Run(WithEntryContext(ctx, entry))
			})
		},
	}, entry.middlewares...)

	// Wrap the job with the middlewares.
	entry.wrappedJob = Chain(middlewares...)(entry.job)

	return entry
}

func (e *Entry) ID() EntryID {
	return e.id
}

// Valid returns true if this is not the zero entry.
func (e *Entry) Valid() bool { return e.id != 0 }

func (e *Entry) Schedule() Schedule {
	return e.schedule
}

func (e *Entry) Next() time.Time {
	return e.next
}

func (e *Entry) Prev() time.Time {
	return e.prev
}

func (e *Entry) WrappedJob() Job {
	return e.wrappedJob
}

func (e *Entry) Job() Job {
	return e.job
}

// ------------------------------------ Entry Context ------------------------------------

type entryContextKey struct{}

// WithEntryContext returns a new context with the given EntryID.
func WithEntryContext(ctx context.Context, entry *Entry) context.Context {
	return context.WithValue(ctx, entryContextKey{}, entry)
}

// EntryFromContext returns the EntryID from the context.
func EntryFromContext(ctx context.Context) (*Entry, bool) {
	entry, ok := ctx.Value(entryContextKey{}).(*Entry)
	return entry, ok
}
