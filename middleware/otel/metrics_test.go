package otel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

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
}
