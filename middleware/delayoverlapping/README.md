# Delay Overlapping Middleware

This middleware is used to delay overlapping cron job runs.

If the previous job is not finished, the next job will be delayed until the previous job is finished.

If the previous job greater than the reminder time, the logger will print the info message.

## Behavior

`delayoverlapping` serializes overlapping runs of the same job with an in-memory mutex.
It does not skip overlapping runs. Instead, each overlapping run waits for the previous
run to finish and then executes.

This means high-frequency schedules combined with long-running jobs can create a backlog
of waiting goroutines. Use this middleware when every scheduled run should eventually
execute in order.

If overlapping runs should be skipped instead of queued, use
[`nooverlapping`](../nooverlapping).

`WithReminderTime` controls when a delayed run is logged as slow. It does not set a job
timeout, cancel waiting runs, or limit the backlog size.

## Usage

```go
package main

import (
	"context"
	"time"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/middleware/delayoverlapping/v4"
)

func main() {
	c := cron.New()
	c.Use(delayoverlapping.New(
		delayoverlapping.WithLogger(cron.DefaultLogger),  // if not set, use cron.DefaultLogger
		delayoverlapping.WithReminderTime(5*time.Minute), // if not set, use 1 minute
	))

	_, _ = c.AddFunc("* * * * *", func(ctx context.Context) error {
		// do something
		return nil
	})

	c.Start()
	defer c.Stop()

	time.Sleep(10 * time.Second)
}
```
