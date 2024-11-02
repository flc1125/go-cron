package main

import (
	"context"
	"log"
	"time"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/v4/middleware/nooverlapping"
	"github.com/flc1125/go-cron/v4/middleware/recovery"
)

func main() {
	c := cron.New(
		cron.WithSeconds(),
		cron.WithMiddleware(
			recovery.New(),
		),
	)

	c.AddFunc("* * * * * *", func(ctx context.Context) error {
		log.Println("hello")
		time.Sleep(time.Second * 3)
		return nil
	})
	c.AddFunc("* * * * * *", func(ctx context.Context) error {
		log.Println("world")
		time.Sleep(time.Second * 3)
		return nil
	}, nooverlapping.New())

	c.Start()
	defer c.Stop()

	time.Sleep(time.Second * 10)
}
