package cron

import (
	"context"
	"time"
)

// EntryID identifies an [Entry] within a [Cron] instance.
type EntryID int

// Entry consists of a [Schedule] and the [Job] to execute on that schedule.
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

// EntryOption configures an [Entry].
type EntryOption func(*Entry)

// WithEntryMiddlewares configures the [Middleware] applied to an [Entry].
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
				runEntry := entry
				if snapshot, ok := executionEntryFromContext(ctx, entry); ok {
					runEntry = snapshot
				}
				return job.Run(WithEntryContext(ctx, runEntry))
			})
		},
	}, entry.middlewares...)

	// Wrap the job with the middlewares.
	entry.wrappedJob = Chain(middlewares...)(entry.job)

	return entry
}

// ID returns the [EntryID] assigned by [Cron].
func (e *Entry) ID() EntryID {
	return e.id
}

// Valid returns true if this is not the zero entry.
func (e *Entry) Valid() bool { return e.id != 0 }

// Schedule returns the [Schedule] associated with the entry.
func (e *Entry) Schedule() Schedule {
	return e.schedule
}

// Next returns the next scheduled activation time.
func (e *Entry) Next() time.Time {
	return e.next
}

// Prev returns the previous scheduled activation time.
func (e *Entry) Prev() time.Time {
	return e.prev
}

// WrappedJob returns the [Job] after its [Middleware] has been applied.
func (e *Entry) WrappedJob() Job {
	return e.wrappedJob
}

// Job returns the originally registered [Job].
func (e *Entry) Job() Job {
	return e.job
}

// ------------------------------------ Entry Context ------------------------------------

type entryContextKey struct{}

type executionEntryContextKey struct{}

type executionEntryContextValue struct {
	registered *Entry
	snapshot   *Entry
}

func withExecutionEntryContext(ctx context.Context, registered, snapshot *Entry) context.Context {
	return context.WithValue(ctx, executionEntryContextKey{}, executionEntryContextValue{
		registered: registered,
		snapshot:   snapshot,
	})
}

func executionEntryFromContext(ctx context.Context, registered *Entry) (*Entry, bool) {
	value, ok := ctx.Value(executionEntryContextKey{}).(executionEntryContextValue)
	if !ok || value.registered != registered {
		return nil, false
	}
	return value.snapshot, true
}

// WithEntryContext returns a new [context.Context] containing the given [Entry].
func WithEntryContext(ctx context.Context, entry *Entry) context.Context {
	return context.WithValue(ctx, entryContextKey{}, entry)
}

// EntryFromContext returns the [Entry] associated with the current job
// execution. Jobs started by [Cron] receive a stable snapshot for that
// execution. Direct [Entry.WrappedJob] calls without a scheduler execution
// context receive the registered [Entry]. In a scheduler execution snapshot,
// [Entry.Prev] is the scheduled activation time for that execution and
// [Entry.Next] is the next scheduled activation time already calculated by the
// scheduler.
func EntryFromContext(ctx context.Context) (*Entry, bool) {
	entry, ok := ctx.Value(entryContextKey{}).(*Entry)
	return entry, ok
}
