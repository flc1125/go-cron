// Package cron provides a cron job scheduler for Go applications.
//
// This package is a fork of robfig/cron with significant improvements including
// context support, a middleware system, and enhanced scheduling capabilities.
//
// # Basic Usage
//
// Creating and starting a basic cron scheduler:
//
//	c := cron.New()
//	c.AddFunc("30 * * * *", func(ctx context.Context) error {
//		fmt.Println("Every hour on the half hour")
//		return nil
//	})
//	c.Start()
//
// # Schedule Format
//
// The scheduler supports both standard 5-field cron expressions and extended
// 6-field expressions with seconds:
//
//	Standard (5-field):  "30 * * * *"        (every hour at 30 minutes)
//	Extended (6-field):  "0 30 * * * *"      (every hour at 30 minutes, 0 seconds)
//
// Field order: [second] minute hour day month weekday
//
// Special characters:
//   - * : matches any value
//   - , : value list separator
//   - - : range of values
//   - / : step values
//   - @yearly, @annually, @monthly, @weekly, @daily, @midnight, @hourly : predefined schedules
//
// # Jobs and Functions
//
// Jobs can be added in multiple ways:
//
//	// Add a function
//	c.AddFunc("* * * * *", func(ctx context.Context) error {
//		// Job implementation
//		return nil
//	})
//
//	// Add a Job interface implementation
//	c.AddJob("* * * * *", cron.JobFunc(func(ctx context.Context) error {
//		// Job implementation
//		return nil
//	}))
//
//	// Custom Job implementation
//	type MyJob struct{}
//	func (j MyJob) Run(ctx context.Context) error {
//		// Job implementation
//		return nil
//	}
//	c.AddJob("* * * * *", MyJob{})
//
// # Middleware System
//
// The package provides a flexible middleware system for job execution:
//
//	c := cron.New(
//		cron.WithMiddleware(
//			recovery.New(), // Recover from panics
//		),
//	)
//
//	// Apply middleware to specific jobs
//	c.AddFunc("* * * * *", myFunc, nooverlapping.New())
//
//	// Apply middleware to jobs registered after this call
//	c.Use(nooverlapping.New())
//
// Available middleware:
//   - recovery: Recovers from panics in job execution
//   - nooverlapping: Prevents concurrent execution of the same job
//   - delayoverlapping: Delays overlapping job execution
//   - distributednooverlapping: Distributed job overlap prevention
//   - otel: OpenTelemetry integration for tracing
//
// # Configuration Options
//
// The cron scheduler can be configured with various options:
//
//	c := cron.New(
//		cron.WithSeconds(),                    // Enable seconds field
//		cron.WithLocation(time.UTC),           // Set timezone
//		cron.WithContext(ctx),                 // Set context
//		cron.WithLogger(myLogger),             // Custom logger
//		cron.WithParser(myParser),             // Custom parser
//	)
//
// # Context Support
//
// All jobs receive a context.Context parameter, enabling:
//   - Cancellation and timeout handling
//   - Value passing between middleware and jobs
//   - Graceful shutdown coordination
//
// [Cron.Stop] does not cancel this context. To cancel running jobs during
// shutdown, pass a cancellable context with [WithContext] and cancel it
// explicitly.
//
// # Lifecycle Management
//
//	c := cron.New()
//
//	// Add jobs
//	entryID, err := c.AddFunc("* * * * *", myFunc)
//	if err != nil {
//		// Handle error
//	}
//
//	// Start the scheduler
//	c.Start()
//
//	// Remove jobs
//	c.Remove(entryID)
//
//	// Stop future scheduling and wait for jobs started by this run
//	<-c.Stop().Done()
//
// [Cron.Stop] does not cancel jobs that have already started. The [Cron] may be
// started again before the context returned by the previous [Cron.Stop] is
// done; jobs from the two runs may overlap, and each returned context waits
// only for jobs from its own run.
//
//	// Get current entries
//	entries := c.Entries()
//
// # Error Handling
//
// The scheduler ignores errors returned by [Job.Run]. [WithLogger] configures
// scheduler messages and does not handle job errors. Handle errors inside the
// [Job] or with [Middleware]:
//
//	c.AddFunc("* * * * *", func(ctx context.Context) error {
//		if err := doSomething(); err != nil {
//			recordJobError(err)
//			return fmt.Errorf("job failed: %w", err)
//		}
//		return nil
//	})
//
// Panics are not recovered by default. Configure the recovery middleware
// explicitly when panic recovery is required.
//
// # Thread Safety
//
// [Cron] is thread-safe and can be safely accessed from multiple goroutines.
// All public methods are protected by appropriate synchronization.
package cron
