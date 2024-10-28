# Cron

![Supported Go Versions](https://img.shields.io/badge/Go-%3E%3D1.18-blue)
[![Package Version](https://badgen.net/github/release/flc1125/go-cron/stable)](https://github.com/flc1125/go-cron/releases)
[![GoDoc](https://pkg.go.dev/badge/github.com/flc1125/go-cron/v4)](https://pkg.go.dev/github.com/flc1125/go-cron/v4)
[![codecov](https://codecov.io/gh/flc1125/go-cron/graph/badge.svg?token=mXNvrv22JH)](https://codecov.io/gh/flc1125/go-cron)
[![Go Report Card](https://goreportcard.com/badge/github.com/flc1125/go-cron)](https://goreportcard.com/report/github.com/flc1125/go-cron)
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

```gopackage main

import (
	"context"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/v4/middleware/nooverlapping"
	"github.com/flc1125/go-cron/v4/middleware/recovery"
)

func main() {
	c := cron.New(
		cron.WithSeconds(), // if you want to use seconds, you can use this option
		cron.WithMiddleware(
			recovery.New(),      // recover panic
			nooverlapping.New(), // prevent job overlapping
		),
		// ... other options
	)

	// use middleware
	c.Use(recovery.New(), nooverlapping.New()) // use middleware

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
	})

	// start cron
	c.Start()

	// stop cron
	c.Stop()
}
```

## License

The MIT License (MIT). Please see [License File](LICENSE) for more information.
