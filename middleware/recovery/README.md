# Recovery Middleware

The `recovery` middleware is a middleware for [go-cron](https://github.com/flc1125/go-cron) that recovers from panics.

## Usage

```go
package main

import (
	"context"
	"time"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/v4/middleware/recovery"
)

func main() {
	c := cron.New()
	c.Use(recovery.New(
		recovery.WithLogger(cron.DefaultLogger), // if not set, use cron.DefaultLogger
	))

	_, _ = c.AddFunc("* * * * * *", func(context.Context) error {
		panic("YOLO")
	})

	c.Start()
	defer c.Stop()

	time.Sleep(2 * time.Second)
}

```