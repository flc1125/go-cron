package recovery

import (
	"context"
	"fmt"
	"runtime"

	"github.com/flc1125/go-cron/v4"
)

func New(logger cron.Logger) cron.Middleware {
	return func(next cron.Job) cron.Job {
		return cron.JobFunc(func(ctx context.Context) {
			defer func() {
				if r := recover(); r != nil {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					logger.Error(err, "panic", "stack", "...\n"+string(buf))
				}
			}()
			next.Run(ctx)
		})
	}
}
