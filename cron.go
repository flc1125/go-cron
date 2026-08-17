package cron

import (
	"context"
	"sort"
	"sync"
	"time"
)

// Cron keeps track of any number of entries, invoking the associated func as
// specified by the schedule. It may be started, stopped, and the entries may
// be inspected while running.
type Cron struct {
	ctx           context.Context
	entries       []*Entry
	middlewares   []Middleware
	middlewaresMu sync.Mutex
	add           chan *Entry
	remove        chan EntryID
	snapshot      chan chan []Entry
	runningMu     sync.Mutex
	running       bool
	runState      *runState
	logger        Logger
	location      *time.Location
	parser        ScheduleParser
	nextID        EntryID
}

type runState struct {
	stop chan struct{}
	jobs sync.WaitGroup
}

// ScheduleParser is an interface for schedule spec parsers that return a Schedule
type ScheduleParser interface {
	Parse(spec string) (Schedule, error)
}

// Schedule describes a job's duty cycle.
type Schedule interface {
	// Next returns the next activation time, later than the given time.
	// Next is invoked initially, and then each time the job is run.
	Next(time.Time) time.Time
}

// byTime is a wrapper for sorting the entry array by time
// (with zero time at the end).
type byTime []*Entry

func (s byTime) Len() int      { return len(s) }
func (s byTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byTime) Less(i, j int) bool {
	// Two zero times should return false.
	// Otherwise, zero is "greater" than any other time.
	// (To sort it at the end of the list.)
	if s[i].next.IsZero() {
		return false
	}
	if s[j].next.IsZero() {
		return true
	}
	return s[i].next.Before(s[j].next)
}

// New returns a new Cron job runner, modified by the given options.
//
// Available Settings
//
//	Time Zone
//	  Description: The time zone in which schedules are interpreted
//	  Default:     time.Local
//
//	Parser
//	  Description: Parser converts cron spec strings into cron.Schedules.
//	  Default:     Accepts this spec: https://en.wikipedia.org/wiki/Cron
//
//	Chain
//	  Description: Wrap submitted jobs to customize behavior.
//	  Default:     A chain that recovers panics and logs them to stderr.
//
// See "cron.With*" to modify the default behavior.
func New(opts ...Option) *Cron {
	c := &Cron{
		ctx:       context.Background(),
		entries:   nil,
		add:       make(chan *Entry),
		snapshot:  make(chan chan []Entry),
		remove:    make(chan EntryID),
		running:   false,
		runningMu: sync.Mutex{},
		logger:    DefaultLogger,
		location:  time.Local,
		parser:    standardParser,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Use appends middleware to the chain of jobs registered after Use returns.
// It does not affect previously registered jobs and is safe to call concurrently
// with job registration.
func (c *Cron) Use(middleware ...Middleware) {
	c.middlewaresMu.Lock()
	defer c.middlewaresMu.Unlock()

	c.middlewares = append(c.middlewares, middleware...)
}

// AddFunc adds a func to the Cron to be run on the given schedule.
// The spec is parsed using the time zone of this Cron instance as the default.
// An opaque id is returned that can be used to later remove it.
func (c *Cron) AddFunc(spec string, cmd func(ctx context.Context) error, middlewares ...Middleware) (EntryID, error) {
	return c.AddJob(spec, JobFunc(cmd), middlewares...)
}

// AddJob adds a Job to the Cron to be run on the given schedule.
// The spec is parsed using the time zone of this Cron instance as the default.
// An opaque id is returned that can be used to later remove it.
func (c *Cron) AddJob(spec string, cmd Job, middlewares ...Middleware) (EntryID, error) {
	schedule, err := c.parser.Parse(spec)
	if err != nil {
		return 0, err
	}
	return c.Schedule(schedule, cmd, middlewares...), nil
}

// Schedule adds a Job to the Cron to be run on the given schedule.
// The job is wrapped with the configured Chain.
func (c *Cron) Schedule(schedule Schedule, cmd Job, middlewares ...Middleware) EntryID {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	c.nextID++
	c.middlewaresMu.Lock()
	entryMiddlewares := make([]Middleware, 0, len(c.middlewares)+len(middlewares))
	entryMiddlewares = append(entryMiddlewares, c.middlewares...)
	c.middlewaresMu.Unlock()
	entryMiddlewares = append(entryMiddlewares, middlewares...)
	entry := newEntry(
		c.nextID, schedule, cmd, WithEntryMiddlewares(
			entryMiddlewares...,
		),
	)
	if !c.running {
		c.entries = append(c.entries, entry)
	} else {
		c.add <- entry
	}
	return entry.id
}

// Entries returns a snapshot of the cron entries.
func (c *Cron) Entries() []Entry {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	if c.running {
		replyChan := make(chan []Entry, 1)
		c.snapshot <- replyChan
		return <-replyChan
	}
	return c.entrySnapshot()
}

// Location gets the time zone location
func (c *Cron) Location() *time.Location {
	return c.location
}

// Entry returns a snapshot of the given entry, or nil if it couldn't be found.
func (c *Cron) Entry(id EntryID) Entry {
	for _, entry := range c.Entries() {
		if id == entry.id {
			return entry
		}
	}
	return Entry{}
}

// Remove an entry from being run in the future.
func (c *Cron) Remove(id EntryID) {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	if c.running {
		c.remove <- id
	} else {
		c.removeEntry(id)
	}
}

// Start the cron scheduler in its own goroutine, or no-op if already started.
// Jobs started by each successful call are tracked independently from jobs
// left running by earlier calls.
func (c *Cron) Start() {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	if c.running {
		return
	}
	state := newRunState()
	c.runState = state
	c.running = true
	go c.run(state)
}

// Run the cron scheduler, or no-op if already running. Jobs started by each
// successful call are tracked independently from jobs left running by earlier
// calls.
func (c *Cron) Run() {
	c.runningMu.Lock()
	if c.running {
		c.runningMu.Unlock()
		return
	}
	state := newRunState()
	c.runState = state
	c.running = true
	c.runningMu.Unlock()
	c.run(state)
}

// run the scheduler.. this is private just due to the need to synchronize
// access to the 'running' state variable.
func (c *Cron) run(state *runState) {
	c.logger.Info("start")

	// Figure out the next activation times for each entry.
	now := c.now()
	for _, entry := range c.entries {
		entry.next = entry.schedule.Next(now)
		c.logger.Info("schedule", "now", now, "entry", entry.ID(), "next", entry.next)
	}

	for {
		// Determine the next entry to run.
		sort.Sort(byTime(c.entries))

		var timer *time.Timer
		if len(c.entries) == 0 || c.entries[0].next.IsZero() {
			// If there are no entries yet, just sleep - it still handles new entries
			// and stop requests.
			timer = time.NewTimer(100000 * time.Hour)
		} else {
			timer = time.NewTimer(c.entries[0].next.Sub(now))
		}

		for {
			select {
			case now = <-timer.C:
				now = now.In(c.location)
				c.logger.Info("wake", "now", now)

				// Run every entry whose next time was less than now
				for _, e := range c.entries {
					if e.next.After(now) || e.next.IsZero() {
						break
					}
					e.prev = e.next
					e.next = e.schedule.Next(now)
					executionEntry := *e
					ctx := withExecutionEntryContext(c.ctx, e, &executionEntry)
					state.startJob(ctx, e.WrappedJob())
					c.logger.Info("run", "now", now, "entry", e.ID(), "next", e.next)
				}

			case newEntry := <-c.add:
				timer.Stop()
				now = c.now()
				newEntry.next = newEntry.schedule.Next(now)
				c.entries = append(c.entries, newEntry)
				c.logger.Info("added", "now", now, "entry", newEntry.ID(), "next", newEntry.next)

			case replyChan := <-c.snapshot:
				replyChan <- c.entrySnapshot()
				continue

			case <-state.stop:
				timer.Stop()
				c.logger.Info("stop")
				return

			case id := <-c.remove:
				timer.Stop()
				now = c.now()
				c.removeEntry(id)
				c.logger.Info("removed", "entry", id)
			}

			break
		}
	}
}

func newRunState() *runState {
	return &runState{stop: make(chan struct{})}
}

// startJob runs the given job with the given context in a new goroutine.
func (s *runState) startJob(ctx context.Context, j Job) {
	s.jobs.Go(func() {
		j.Run(ctx) //nolint:errcheck
	})
}

// now returns current time in c location
func (c *Cron) now() time.Time {
	return time.Now().In(c.location)
}

// Stop stops the cron scheduler if it is running. A context is returned so the
// caller can wait for jobs started by that run to complete. If the scheduler is
// not running, the context waits for jobs from the most recent run. It does not
// wait for jobs started by later calls to Start or Run.
func (c *Cron) Stop() context.Context {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	if c.running {
		c.runState.stop <- struct{}{}
		c.running = false
	}
	state := c.runState
	ctx, cancel := context.WithCancel(context.Background())
	if state == nil {
		cancel()
		return ctx
	}
	go func() {
		state.jobs.Wait()
		cancel()
	}()
	return ctx
}

// IsRunning returns true if the cron scheduler is started.
func (c *Cron) IsRunning() bool {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	return c.running
}

// entrySnapshot returns a copy of the current cron entry list.
func (c *Cron) entrySnapshot() []Entry {
	entries := make([]Entry, len(c.entries))
	for i, e := range c.entries {
		entries[i] = *e
	}
	return entries
}

func (c *Cron) removeEntry(id EntryID) {
	var entries []*Entry
	for _, e := range c.entries {
		if e.ID() != id {
			entries = append(entries, e)
		}
	}
	c.entries = entries
}
