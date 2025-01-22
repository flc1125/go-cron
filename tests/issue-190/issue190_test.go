package issue_190 //nolint:revive

import (
	"context"
	"testing"
	"time"

	"github.com/flc1125/go-cron/middleware/distributednooverlapping/redismutex/v4"
	"github.com/flc1125/go-cron/middleware/distributednooverlapping/v4"
	"github.com/flc1125/go-cron/middleware/nooverlapping/v4"
	"github.com/flc1125/go-cron/middleware/otel/v4"
	"github.com/flc1125/go-cron/middleware/recovery/v4"
	"github.com/flc1125/go-cron/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type exampleJob struct {
	t    *testing.T
	name string
}

var (
	_ cron.Job                              = (*exampleJob)(nil)
	_ distributednooverlapping.JobWithMutex = (*exampleJob)(nil)
	_ otel.JobWithName                      = (*exampleJob)(nil)
)

func (j exampleJob) Run(context.Context) error {
	time.Sleep(2 * time.Second)
	j.t.Logf("job %s is running", j.name)
	return nil
}

func (j exampleJob) Name() string {
	return j.name
}

func (j exampleJob) GetMutexKey() string {
	return j.Name()
}

func (j exampleJob) GetMutexTTL() time.Duration {
	return time.Minute
}

func TestIssue190(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.FlushAll(context.Background())

	mutex := redismutex.New(rdb, redismutex.WithPrefix("cron"))
	imsb := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(imsb))

	c := cron.New(
		cron.WithSeconds(),
		cron.WithMiddleware(
			recovery.New(),
			nooverlapping.New(),
			distributednooverlapping.New(mutex),
			otel.New(otel.WithTracerProvider(provider)),
		),
	)

	_, _ = c.AddJob("* * * * * *", &exampleJob{t: t, name: "job1"})

	c.Start()
	defer c.Stop()

	time.Sleep(6 * time.Second)

	assert.LessOrEqual(t, len(imsb.GetSpans()), 3)
}
