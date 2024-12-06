package otel

import (
	"context"
	"testing"

	"github.com/flc1125/go-cron/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

			require.Equal(t, tt.error, middleware(&mockJob{t: t, name: tt.name, err: tt.error}).Run(ctx))
			require.Len(t, imsb.GetSpans(), 1)

			span := imsb.GetSpans()[0]
			assert.Equal(t, "cron "+tt.name, span.Name)
			assert.NotEmpty(t, span.SpanContext.TraceID())
			assert.NotEmpty(t, span.SpanContext.SpanID())
			assert.Equal(t, trace.SpanKindInternal, span.SpanKind)
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
