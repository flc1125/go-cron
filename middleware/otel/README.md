# OTel Middleware

The `otel` middleware provides OpenTelemetry traces and metrics for cron Job
executions.

> [!WARNING]  
> **Unstable Semantic Conventions**
>
> OpenTelemetry has not yet defined semantic conventions that align with cron job scheduling and execution. As a result, all metrics, attributes, and trace semantics provided by this middleware are custom-defined and subject to change.
>
> These conventions will remain unstable until OTel releases official semantic conventions for cron-like workloads. When that happens, this middleware will be updated to adopt the official conventions, which **will be a breaking change** that may require updates to your dashboards, queries, and alerting rules.

## Behavior

Scheduler Jobs are instrumented only when the registered Job implements
`JobWithName`. Calls without an Entry context and Jobs without a name run without
cron trace or metric data.

When an instrumented Job returns an error, the middleware records it on the span
and metrics, then returns the error unchanged. The core scheduler does not log or
otherwise handle that returned error, so applications still need Job-level or
additional middleware error handling.

The `cron.job.prev.time` and `cron.job.next.time` span attributes come from the
stable Entry snapshot for that execution. They represent the current scheduled
activation time and the next scheduled activation time, respectively.

Middleware order is outermost to innermost. To recover panics raised by OTel or
the Job, register recovery before OTel, for example
`cron.WithMiddleware(recovery.New(), otel.New())`.

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
	time.Sleep(10 * time.Second)
	<-c.Stop().Done()
	fmt.Println("spans:", len(imsb.GetSpans()))
}
```

output:

```shell
spans: 10
```
