package main

import (
	"context"
	"time"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/v4/middleware/nooverlapping"
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
