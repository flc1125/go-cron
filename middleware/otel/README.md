# OTel Middleware

The `otel` is a middleware for that provides observability with OpenTelemetry.

> [!WARNING]  
> **Unstable Semantic Conventions**
>
> OpenTelemetry has not yet defined semantic conventions that align with cron job scheduling and execution. As a result, all metrics, attributes, and trace semantics provided by this middleware are custom-defined and subject to change.
>
> These conventions will remain unstable until OTel releases official semantic conventions for cron-like workloads. When that happens, this middleware will be updated to adopt the official conventions, which **will be a breaking change** that may require updates to your dashboards, queries, and alerting rules.

## Usage

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/middleware/otel/v4"
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