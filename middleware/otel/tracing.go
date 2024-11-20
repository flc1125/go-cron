package otel

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/flc1125/go-cron/v4"
)

const ScopeName = "github.com/flc1125/go-cron/v4/middleware/otel"

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
	tracer := o.tp.Tracer(ScopeName)
	return func(original cron.Job) cron.Job {
		return cron.JobFunc(func(ctx context.Context) error {
			job, ok := any(original).(JobWithName)
			if !ok {
				return original.Run(ctx)
			}

			ctx, span := tracer.Start(ctx, "cron "+job.Name(),
				trace.WithSpanKind(trace.SpanKindInternal),
			)
			defer span.End()

			span.SetAttributes(append(
				entryAttributes(ctx),
				attrJobName.String(job.Name()),
			)...)

			err := job.Run(ctx)

			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
			} else {
				span.SetStatus(codes.Ok, "OK")
			}

			return err
		})
	}
}

func entryAttributes(ctx context.Context) []attribute.KeyValue {
	entry, ok := cron.EntryFromContext(ctx)
	if !ok {
		return []attribute.KeyValue{}
	}

	return []attribute.KeyValue{
		attrJobID.Int(int(entry.ID())),
		attrJobPrevTime.String(entry.Prev().String()),
		attrJobNextTime.String(entry.Next().String()),
	}
}
