# Recovery Middleware

The `recovery` middleware is a middleware for [go-cron](https://github.com/flc1125/go-cron) that recovers from panics.

## Usage

```go
package recovery_test

import (
	"context"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/v4/middleware/recovery"
)

func Example() {
	c := cron.New()
	c.Use(recovery.New())

	c.AddFunc("* * * * * ?", func(ctx context.Context) error {
		panic("YOLO")
	})

	c.Start()
	defer c.Stop()
}
```