# No Overlapping Middleware

This middleware is used to prevent the overlapping of the cron job.

If the previous job is not finished, the next job will be skipped and the logger will print the info message.

## Usage

```go
package main

import (
	"context"
	"time"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/middleware/nooverlapping/v4"
)

func main() {

	c := cron.New()
	c.Use(nooverlapping.New(
		nooverlapping.WithLogger(cron.DefaultLogger), // if not set, use cron.DefaultLogger
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