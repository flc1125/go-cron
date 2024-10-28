# Delay Overlapping Middleware

This middleware is used to delay the overlapping of the cron job. 

If the previous job is not finished, the next job will be delayed until the previous job is finished.

If the previous job greater than the reminder time, the logger will print the info message.

## Usage

```go
package main

import (
	"context"
	"time"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/v4/middleware/delayoverlapping"
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