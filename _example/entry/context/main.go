package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/flc1125/go-cron/v4"
)

func main() {
	c := cron.New(
		cron.WithSeconds(),
	)

	_, _ = c.AddFunc("* * * * * *", func(ctx context.Context) error {
		entry, ok := cron.EntryFromContext(ctx)
		if ok {
			log.Println(fmt.Sprintf("entry id: %d", entry.ID))
		}
		return nil
	})

	c.Start()
	defer c.Stop()

	time.Sleep(5 * time.Second)
}
