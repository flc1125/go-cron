// Package cron provides a cron job scheduler for Go applications.
//
// This package is a fork of robfig/cron with significant improvements including
// context support, middleware system, better error handling, and enhanced
// scheduling capabilities.
//
// # Basic Usage
//
// Creating and starting a basic cron scheduler:
//
//	c := cron.New()
//	c.AddFunc("0 30 * * * *", func(ctx context.Context) error {
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
//			recovery.New(),      // Recover from panics
//			nooverlapping.New(), // Prevent job overlapping
//		),
//	)
//
//	// Apply middleware to specific jobs
//	c.AddFunc("* * * * *", myFunc, recovery.New())
//
//	// Use middleware on running cron
//	c.Use(recovery.New(), nooverlapping.New())
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
//	// Stop the scheduler (waits for running jobs)
//	c.Stop()
//
//	// Get current entries
//	entries := c.Entries()
//
// # Error Handling
//
// Jobs can return errors, which are handled by the configured logger:
//
//	c.AddFunc("* * * * *", func(ctx context.Context) error {
//		if err := doSomething(); err != nil {
//			return fmt.Errorf("job failed: %w", err)
//		}
//		return nil
//	})
//
// # Thread Safety
//
// The cron scheduler is thread-safe and can be safely accessed from multiple
// goroutines. All public methods are protected by appropriate synchronization.
package cron
