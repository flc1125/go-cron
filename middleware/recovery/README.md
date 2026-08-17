# Recovery Middleware

The `recovery` middleware recovers panics raised by the Job or middleware it
wraps and logs the panic and stack trace. Panic recovery is not enabled by
default; register this middleware explicitly when it is required.

Returned Job errors pass through unchanged. When a panic is recovered, that
execution returns a nil error after the panic has been logged.

Middleware order is outermost to innermost. Register recovery before any
middleware whose panics it should catch. For example,
`cron.WithMiddleware(recovery.New(), other)` produces
`recovery(other(job))`.

## Usage

```go
package main

import (
	"context"
	"time"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/middleware/recovery/v4"
)

func main() {
	c := cron.New(
		cron.WithSeconds(),
		cron.WithMiddleware(
			recovery.New(
				recovery.WithLogger(cron.DefaultLogger), // default: cron.DefaultLogger
			),
		),
	)

	_, _ = c.AddFunc("* * * * * *", func(context.Context) error {
		panic("YOLO")
	})

	c.Start()
	time.Sleep(2 * time.Second)
	<-c.Stop().Done()
}
```
