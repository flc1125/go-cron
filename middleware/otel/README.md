# OTel Middleware

The `otel` is a middleware for that provides observability with OpenTelemetry.

## Usage

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/v4/middleware/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type basicJob struct {
}

var (
	_ cron.Job         = (*basicJob)(nil)
	_ otel.JobWithName = (*basicJob)(nil)
)

func (b *basicJob) Name() string {
	return "basic:job"
}

func (b *basicJob) Run(ctx context.Context) error {
	// do something
	return nil
}

func main() {
	// configure otel, the following is just a demonstration provider.
	imsb := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(imsb))

	// cron
	c := cron.New(cron.WithSeconds())
	c.Use(otel.New(
		otel.WithTracerProvider(tp), // custom otel.TracerProvider
	))

	_, _ = c.AddJob("* * * * * *", &basicJob{})

	c.Start()
	defer c.Stop()

	time.Sleep(10 * time.Second)
	fmt.Println("spans:", len(imsb.GetSpans()))
}
```

output:

```shell
spans: 10
```