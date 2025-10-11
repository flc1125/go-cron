package otel

import (
	"context"
	"testing"

	"github.com/flc1125/go-cron/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/metric/metricdata/metricdatatest"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

var (
	imsb       = tracetest.NewInMemoryExporter()
	provider   = sdktrace.NewTracerProvider(sdktrace.WithSyncer(imsb))
	ctx        = context.Background()
	middleware = New(WithTracerProvider(provider))
)

type mockJob struct {
	t    *testing.T
	name string
	err  error
}

func (m *mockJob) Name() string {
	return m.name
}

func (m *mockJob) Run(ctx context.Context) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("test.job", m.name))
	return m.err
}

func TestTracing(t *testing.T) {
	tests := []struct {
		name         string
		error        error
		extraTesting func(t *testing.T, span tracetest.SpanStub)
	}{
		{"test success", nil, func(t *testing.T, span tracetest.SpanStub) {
			assert.Equal(t, codes.Unset, span.Status.Code)
		}},
		{"test error", assert.AnError, func(t *testing.T, span tracetest.SpanStub) {
			assert.Equal(t, codes.Error, span.Status.Code)

			require.Len(t, span.Events, 1)
			event := span.Events[0]
			assert.Equal(t, "exception", event.Name)
			assert.Contains(t, event.Attributes, attribute.String("exception.type", "*errors.errorString"))
			assert.Contains(t, event.Attributes, attribute.String("exception.message", assert.AnError.Error()))
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer imsb.Reset()

			// Create a cron instance with the middleware
			c := cron.New(cron.WithMiddleware(middleware))

			// Add the job
			entryID, err := c.AddJob("@every 1h", &mockJob{t: t, name: tt.name, err: tt.error})
			require.NoError(t, err)

			// Get the entry to access the wrapped job
			entry := c.Entry(entryID)
			require.True(t, entry.Valid())

			require.Equal(t, tt.error, entry.WrappedJob().Run(ctx))
			require.Len(t, imsb.GetSpans(), 1)

			span := imsb.GetSpans()[0]
			assert.Equal(t, "cron "+tt.name, span.Name)
			assert.NotEmpty(t, span.SpanContext.TraceID())
			assert.NotEmpty(t, span.SpanContext.SpanID())
			assert.Equal(t, trace.SpanKindInternal, span.SpanKind)
			assert.Contains(t, span.Attributes, attribute.Int("cron.job.id", int(entry.ID())))
			assert.Contains(t, span.Attributes, attribute.String("cron.job.name", tt.name))
			assert.Contains(t, span.Attributes, attribute.String("test.job", tt.name))
			tt.extraTesting(t, span)
		})
	}
}

func TestTracing_NotJobWithName(t *testing.T) {
	defer imsb.Reset()

	require.NoError(t, middleware(cron.JobFunc(func(context.Context) error {
		return nil
	})).Run(ctx))
	require.Len(t, imsb.GetSpans(), 0)
}

func TestMetrics(t *testing.T) {
	r := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(r))
	middleware := New(WithMeterProvider(mp))

	jobName := "metric-job"

	// Create a cron instance with the middleware
	c := cron.New(cron.WithMiddleware(middleware))

	// Add the job
	entryID, err := c.AddJob("@every 1h", &mockJob{t: t, name: jobName, err: nil})
	require.NoError(t, err)

	// Get the entry to access the wrapped job
	entry := c.Entry(entryID)
	require.True(t, entry.Valid())

	require.NoError(t, entry.WrappedJob().Run(t.Context()))

	var rm metricdata.ResourceMetrics
	require.NoError(t, r.Collect(ctx, &rm))
	assert.Len(t, rm.ScopeMetrics, 1)

	// assert scope
	assert.Equal(t, instrumentation.Scope{
		Name:    "github.com/flc1125/go-cron/middleware/otel/v4",
		Version: cron.Version(),
	}, rm.ScopeMetrics[0].Scope)

	// assert counter metric
	metricdatatest.AssertEqual(t, metricdata.Metrics{
		Name:        "cron.job.run.counter",
		Description: "The total number of cron job runs",
		Unit:        "{data_point}",
		Data: metricdata.Sum[int64]{
			Temporality: metricdata.CumulativeTemporality,
			DataPoints: []metricdata.DataPoint[int64]{
				{
					Attributes: attribute.NewSet(
						attribute.Int64("cron.job.id", int64(entry.ID())),
						attribute.String("cron.job.name", jobName),
					),
					Value: 1,
				},
			},
			IsMonotonic: true,
		},
	}, rm.ScopeMetrics[0].Metrics[0], metricdatatest.IgnoreTimestamp())

	// assert inflight metric
	metricdatatest.AssertEqual(t, metricdata.Metrics{
		Name:        "cron.job.run.inflight",
		Description: "The number of cron jobs currently running",
		Unit:        "{data_point}",
		Data: metricdata.Sum[int64]{
			Temporality: metricdata.CumulativeTemporality,
			DataPoints: []metricdata.DataPoint[int64]{
				{
					Attributes: attribute.NewSet(
						attribute.Int64("cron.job.id", int64(entry.ID())),
						attribute.String("cron.job.name", jobName),
					),
					Value: 0,
				},
			},
			IsMonotonic: false,
		},
	}, rm.ScopeMetrics[0].Metrics[1], metricdatatest.IgnoreTimestamp())

	// assert duration metric
	metricdatatest.AssertEqual(t, metricdata.Metrics{
		Name:        "cron.job.run.duration",
		Description: "The duration of cron job runs",
		Unit:        "s",
		Data: metricdata.Histogram[float64]{
			Temporality: metricdata.CumulativeTemporality,
			DataPoints: []metricdata.HistogramDataPoint[float64]{
				{
					Attributes: attribute.NewSet(
						attribute.Int64("cron.job.id", int64(entry.ID())),
						attribute.String("cron.job.name", jobName),
					),
					Count: 1,
				},
			},
		},
	}, rm.ScopeMetrics[0].Metrics[2], metricdatatest.IgnoreTimestamp(), metricdatatest.IgnoreValue())
}
