package otel

import (
	"context"

	"github.com/flc1125/go-cron/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const scopeName = "github.com/flc1125/go-cron/middleware/otel/v4"

var (
	attrJobName     = attribute.Key("cron.job.name")
	attrJobID       = attribute.Key("cron.job.id")
	attrJobPrevTime = attribute.Key("cron.job.prev.time")
	attrJobNextTime = attribute.Key("cron.job.next.time")
)

type options struct {
	tp trace.TracerProvider
}

type Option func(*options)

func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(o *options) {
		o.tp = tp
	}
}

func newOption(opts ...Option) *options {
	opt := &options{
		tp: otel.GetTracerProvider(),
	}
	for _, o := range opts {
		o(opt)
	}
	return opt
}

type JobWithName interface {
	cron.Job

	// Name returns the name of the job.
	Name() string
}

func New(opts ...Option) cron.Middleware {
	o := newOption(opts...)
	tracer := o.tp.Tracer(scopeName)
	return func(original cron.Job) cron.Job {
		return cron.JobFunc(func(ctx context.Context) error {
			entry, ok := cron.EntryFromContext(ctx)
			if !ok {
				return original.Run(ctx)
			}

			job, ok := any(entry.Job()).(JobWithName)
			if !ok {
				return original.Run(ctx)
			}

			ctx, span := tracer.Start(ctx, "cron "+job.Name(),
				trace.WithSpanKind(trace.SpanKindInternal),
			)
			defer span.End()

			span.SetAttributes(
				attrJobID.Int(int(entry.ID())),
				attrJobName.String(job.Name()),
				attrJobPrevTime.String(entry.Prev().String()),
				attrJobNextTime.String(entry.Next().String()),
			)

			err := job.Run(ctx)
			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
			}

			return err
		})
	}
}
