package recovery_test

import (
	"context"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/v4/middleware/recovery"
)

func Example() {
	c := cron.New()
	c.Use(recovery.New())

	_, _ = c.AddFunc("* * * * * ?", func(context.Context) error {
		panic("YOLO")
	})

	c.Start()
	defer c.Stop()
}
