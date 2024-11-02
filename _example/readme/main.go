package main

import (
	"context"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/v4/middleware/nooverlapping"
	"github.com/flc1125/go-cron/v4/middleware/recovery"
)

func main() {
	c := cron.New(
		cron.WithSeconds(), // if you want to use seconds, you can use this option
		cron.WithMiddleware(
			recovery.New(),      // recover panic
			nooverlapping.New(), // prevent job overlapping
		),
		cron.WithContext(context.Background()), // use custom context
		// ... other options
	)

	// use middleware
	c.Use(recovery.New(), nooverlapping.New()) // use middleware

	// add job
	entryID, _ := c.AddJob("* * * * * *", cron.JobFunc(func(ctx context.Context) error {
		// do something
		return nil
	}))
	_ = entryID

	// add func
	_, _ = c.AddFunc("* * * * * *", func(ctx context.Context) error {
		// do something
		return nil
	}, recovery.New(), nooverlapping.New()) // use middleware for this job

	// start cron
	c.Start()

	// stop cron
	c.Stop()
}
