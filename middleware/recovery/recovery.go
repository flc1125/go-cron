package recovery

import (
	"context"
	"fmt"
	"runtime"

	"github.com/flc1125/go-cron/v4"
)

const size = 64 << 10

type options struct {
	logger cron.Logger
}

type Option func(*options)

func newOptions(opts ...Option) *options {
	opt := &options{
		logger: cron.DefaultLogger,
	}
	for _, o := range opts {
		o(opt)
	}
	return opt
}

func WithLogger(logger cron.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

// New returns a new recovery middleware.
// It recovers from any panics and logs the panic with the provided logger.
func New(opts ...Option) cron.Middleware {
	o := newOptions(opts...)
	return func(next cron.Job) cron.Job {
		return cron.JobFunc(func(ctx context.Context) error {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					o.logger.Error(err, "panic", "stack", "...\n"+string(buf))
				}
			}()
			return next.Run(ctx)
		})
	}
}
