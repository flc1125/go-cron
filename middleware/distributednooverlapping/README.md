# Distributed No Overlapping Middleware

This middleware is used to prevent the overlapping of the cron job in a distributed environment.

## Features

- Distributed locking using Redis
- Configurable mutex TTL and key prefixes
- Automatic lock release on job completion
- Graceful handling of network partitions

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/middleware/distributednooverlapping/v4"
	"github.com/flc1125/go-cron/middleware/distributednooverlapping/redismutex/v4"
	"github.com/redis/go-redis/v9"
)

type basicJob struct {
	slug string
}

var (
	_ cron.Job                              = (*basicJob)(nil)
	_ distributednooverlapping.JobWithMutex = (*basicJob)(nil)
)

func (j *basicJob) GetMutexKey() string {
	return "basic:job"
}

func (j *basicJob) GetMutexTTL() time.Duration {
	return time.Hour * 60 // the ttl suggests greater than the running time of the job
}

func (j *basicJob) Run(ctx context.Context) error {
	time.Sleep(time.Second * 1)
	log.Println(fmt.Sprintf("running job %s", j.slug))
	return nil
}

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	mutex := redismutex.New(rdb, redismutex.WithPrefix("cron"))

	c := cron.New(
		cron.WithSeconds(),
	)
	c.Use(distributednooverlapping.New(mutex,
		distributednooverlapping.WithLogger(cron.DefaultLogger)))

	_, _ = c.AddJob("* * * * * *", &basicJob{"one"})
	_, _ = c.AddJob("* * * * * *", &basicJob{"two"})

	c.Start()
	defer c.Stop()

	time.Sleep(10 * time.Second)
}
```

output:

```shell
2024/12/06 10:35:09 running job two
2024/12/06 10:35:11 running job two
2024/12/06 10:35:13 running job two
2024/12/06 10:35:15 running job one
2024/12/06 10:35:17 running job two
```