# Cron

![Supported Go Versions](https://img.shields.io/badge/Go-%3E%3D1.26.0-blue)
[![Package Version](https://badgen.net/github/release/flc1125/go-cron/stable)](https://github.com/flc1125/go-cron/releases)
[![GoDoc](https://pkg.go.dev/badge/github.com/flc1125/go-cron/v4)](https://pkg.go.dev/github.com/flc1125/go-cron/v4)
[![codecov](https://codecov.io/gh/flc1125/go-cron/graph/badge.svg?token=mXNvrv22JH)](https://codecov.io/gh/flc1125/go-cron)
[![lint](https://github.com/flc1125/go-cron/actions/workflows/lint.yml/badge.svg)](https://github.com/flc1125/go-cron/actions/workflows/lint.yml)
[![tests](https://github.com/flc1125/go-cron/actions/workflows/test.yml/badge.svg)](https://github.com/flc1125/go-cron/actions/workflows/test.yml)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)

The cron library is a cron job library for Go. 

It is a fork of [robfig/cron](https://github.com/robfig/cron) with some improvements.

Thanks to [robfig/cron](https://github.com/robfig/cron) for the original work, and thanks to all the contributors.

> [!IMPORTANT]  
> `v4.x` may introduce situations that are not backward compatible.
>
> The reason for this is that we are using `v4.x` as a transitional version. In this version, we will try to improve the functionality of the components as much as possible until the release of `v5.x`.
>
> When releasing a new version, backward compatibility is the default behavior. If there are any incompatibilities, they will be indicated in the release notes.

## Installation

```bash
go get github.com/flc1125/go-cron/v4
```

## Usage

```go
package main

import (
	"context"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/middleware/nooverlapping/v4"
	"github.com/flc1125/go-cron/middleware/recovery/v4"
)

func main() {
	c := cron.New(
		cron.WithSeconds(), // if you want to use seconds, you can use this option
		cron.WithMiddleware(
			recovery.New(), // recover panic
		),
		cron.WithContext(context.Background()), // use custom context
		// ... other options
	)

	// add job
	entryID, _ := c.AddJob("* * * * * *", cron.JobFunc(func(ctx context.Context) error {
		// do something
		return nil
	}))
	_ = entryID

	// add func
	_, _ = c.AddFunc("* * * * * *", func(ctx context.Context) error {
		// do something
		return nil
	}, nooverlapping.New()) // use middleware for this job

	// start cron
	c.Start()

	// stop future scheduling and wait for jobs started by this run
	<-c.Stop().Done()
}
```

## Middleware

- [recovery](./middleware/recovery): Recovers from panics in job execution, ensuring system stability.
- [delayoverlapping](./middleware/delayoverlapping): Delays execution of overlapping jobs instead of running them concurrently.
- [nooverlapping](./middleware/nooverlapping): Prevents concurrent execution of the same job.
- [distributednooverlapping](./middleware/distributednooverlapping): Prevents concurrent execution across multiple instances using distributed locking.
- [otel](./middleware/otel): Provides OpenTelemetry integration for job execution tracing and metrics.

### Registration and order

`WithMiddleware` configures the initial middleware chain. `Use` appends middleware
only for jobs registered after `Use` returns; it does not modify existing entries.
`Use` is safe to call concurrently with job registration.

Middleware runs in registration order. For example, `Chain(m1, m2)` produces
`m1(m2(job))`. Cron-level middleware configured through `WithMiddleware` or `Use`
runs outside middleware passed to an individual `AddFunc`, `AddJob`, or `Schedule`
call.

## Lifecycle

`Stop` prevents the current scheduler run from starting more jobs. It does not
cancel jobs that have already started. The returned context is done after jobs
started by that run have completed:

```go
c.Start()
// ...
<-c.Stop().Done()
```

A stopped `Cron` may be started again without waiting for the previous Stop
context. Jobs left running by the previous run may overlap jobs started by the
new run. Each Stop context waits only for its own run and is not extended by a
later `Start` or `Run` call.

## Job errors and panics

The scheduler ignores the error returned by `Job.Run`. `WithLogger` configures
scheduler messages; it does not log or otherwise handle Job errors. Handle errors
inside the Job or with middleware. Middleware such as `otel` may observe an error
before returning it to the scheduler.

Panic recovery is not enabled by default. Add the
[`recovery`](./middleware/recovery) middleware explicitly when jobs or downstream
middleware must be recovered.

## Entry context

Jobs started by the scheduler can call `EntryFromContext` to inspect a stable
Entry snapshot for that execution. `Prev` is the scheduled activation time for
the current execution, and `Next` is the next scheduled activation time already
calculated by the scheduler. The scheduler does not mutate that snapshot after
the Job starts.

## License

- The MIT License (MIT). Please see [License File](LICENSE) for more information.
- The original work is licensed under the MIT License (MIT). Please see [robfig/cron](https://github.com/robfig/cron) [License File](https://github.com/robfig/cron/blob/master/LICENSE)
