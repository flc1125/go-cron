package otel

import (
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

type metrics struct {
	counter  metric.Int64Counter
	inflight metric.Int64UpDownCounter
	duration metric.Float64Histogram
}

func newMetrics(
	meter metric.Meter,
) *metrics {
	var err error
	counter, e := meter.Int64Counter(
		"cron.job.run.counter",
		metric.WithDescription("The total number of cron job runs"),
		metric.WithUnit("{data_point}"),
	)
	if e != nil {
		err = errors.Join(err, fmt.Errorf("create counter metric: %w", e))
		counter = noop.Int64Counter{}
	}

	inflight, e := meter.Int64UpDownCounter(
		"cron.job.run.inflight",
		metric.WithDescription("The number of cron jobs currently running"),
		metric.WithUnit("{data_point}"),
	)
	if e != nil {
		err = errors.Join(err, fmt.Errorf("create inflight metric: %w", e))
		inflight = noop.Int64UpDownCounter{}
	}

	duration, e := meter.Float64Histogram(
		"cron.job.run.duration",
		metric.WithDescription("The duration of cron job runs"),
		metric.WithUnit("s"),
	)
	if e != nil {
		err = errors.Join(err, fmt.Errorf("create duration metric: %w", e))
		duration = noop.Float64Histogram{}
	}

	if err != nil {
		otel.Handle(err) // report errors but continue with what we have
	}

	return &metrics{
		counter:  counter,
		inflight: inflight,
		duration: duration,
	}
}
