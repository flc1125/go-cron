package otel

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type failingMeter struct {
	noop.Meter
}

var errInstrument = errors.New("instrument failed")

func (failingMeter) Int64Counter(string, ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	return noop.Int64Counter{}, errInstrument
}

func (failingMeter) Int64UpDownCounter(string, ...metric.Int64UpDownCounterOption) (metric.Int64UpDownCounter, error) {
	return noop.Int64UpDownCounter{}, errInstrument
}

func (failingMeter) Float64Histogram(string, ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
	return noop.Float64Histogram{}, errInstrument
}

func TestMetrics_newMetrics(t *testing.T) {
	t.Run("noop meter", func(t *testing.T) {
		m := newMetrics(noop.NewMeterProvider().Meter("test"))
		assert.NotNil(t, m)
		assert.Equal(t, noop.Int64Counter{}, m.counter)
		assert.Equal(t, noop.Int64UpDownCounter{}, m.inflight)
		assert.Equal(t, noop.Float64Histogram{}, m.duration)
	})

	t.Run("nil meter", func(t *testing.T) {
		m := newMetrics(nil)
		assert.NotNil(t, m)
		assert.Equal(t, noop.Int64Counter{}, m.counter)
		assert.Equal(t, noop.Int64UpDownCounter{}, m.inflight)
		assert.Equal(t, noop.Float64Histogram{}, m.duration)
	})

	t.Run("valid meter", func(t *testing.T) {
		mp := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(sdkmetric.NewManualReader()),
		)
		m := newMetrics(mp.Meter("test"))
		assert.NotNil(t, m)
		assert.NotEqual(t, noop.Int64Counter{}, m.counter)
		assert.NotEqual(t, noop.Int64UpDownCounter{}, m.inflight)
		assert.NotEqual(t, noop.Float64Histogram{}, m.duration)
	})

	t.Run("instrument errors", func(t *testing.T) {
		m := newMetrics(failingMeter{})
		assert.NotNil(t, m)
		assert.Equal(t, noop.Int64Counter{}, m.counter)
		assert.Equal(t, noop.Int64UpDownCounter{}, m.inflight)
		assert.Equal(t, noop.Float64Histogram{}, m.duration)
	})
}
