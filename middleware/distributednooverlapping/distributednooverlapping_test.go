package distributednooverlapping

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/v4/internal/logger"
)

var ctx = context.Background()

type testJob struct {
	cron.Job

	t    *testing.T
	name string
	ttl  time.Duration
}

var _ JobWithMutex = testJob{}

func (j testJob) GetMutexKey() string {
	return j.name
}

func (j testJob) GetMutexTTL() time.Duration {
	return j.ttl
}

type testMutex struct {
	t *testing.T
}

var _ Mutex = testMutex{}

func (m testMutex) Lock(ctx context.Context, job JobWithMutex) (bool, error) {
	if job.GetMutexKey() == "test" {
		return true, nil
	}

	return false, nil
}

func (m testMutex) Unlock(context.Context, JobWithMutex) error {
	return nil
}

func TestMiddleware_Noop(t *testing.T) {
	buffer := logger.NewBuffer()
	ch := make(chan struct{}, 200)
	wg := sync.WaitGroup{}

	noopMiddleware := New(
		NoopMutex{},
		WithLogger(logger.NewBufferLogger(buffer)),
	)

	for i := 0; i < 100; i++ {
		wg.Add(2)

		// not mutex job, so no blocking
		go assert.NoError(t, noopMiddleware(cron.JobFunc(func(ctx context.Context) error {
			defer wg.Done()
			time.Sleep(1 * time.Millisecond)
			ch <- struct{}{}
			return nil
		})).Run(ctx))

		// is mutex job, but noop mutex, so no blocking
		go assert.NoError(t, noopMiddleware(testJob{
			t:    t,
			name: "test",
			ttl:  time.Second,
			Job: cron.JobFunc(func(ctx context.Context) error {
				defer wg.Done()
				time.Sleep(1 * time.Millisecond)
				ch <- struct{}{}
				return nil
			}),
		}).Run(ctx))
	}

	wg.Wait()
	assert.Len(t, ch, 200)
	assert.Empty(t, buffer.String())
}

// func TestMiddleware_Mutex(t *testing.T) {
// 	buffer := logger.NewBuffer()
// 	ch := make(chan struct{}, 200)
// 	wg := sync.WaitGroup{}
//
// 	mutexMiddleware := New(
// 		testMutex{t: t},
// 		WithLogger(logger.NewBufferLogger(buffer)),
// 	)
//
// 	for i := 0; i < 100; i++ {
// 		wg.Add(3)
//
// 		// not mutex job, so no blocking
// 		go assert.NoError(t, mutexMiddleware(cron.JobFunc(func(ctx context.Context) error {
// 			defer wg.Done()
// 			time.Sleep(1 * time.Millisecond)
// 			ch <- struct{}{}
// 			return nil
// 		})).Run(ctx))
//
// 		// mutex job, because the mutex is acquired, but the job is getting the mutex, so no blocking
// 		go assert.NoError(t, mutexMiddleware(testJob{
// 			t:    t,
// 			name: "test",
// 			ttl:  time.Second * 2,
// 			Job: cron.JobFunc(func(ctx context.Context) error {
// 				defer wg.Done()
// 				time.Sleep(10 * time.Millisecond)
// 				ch <- struct{}{}
// 				return nil
// 			}),
// 		}).Run(ctx))
//
// 		// mutex job, because the mutex is acquired, but the job is not getting the mutex, so blocking
// 		go assert.NoError(t, mutexMiddleware(testJob{
// 			t:    t,
// 			name: "test111",
// 			ttl:  time.Second * 2,
// 			Job: cron.JobFunc(func(ctx context.Context) error {
// 				defer wg.Done()
// 				time.Sleep(10 * time.Millisecond)
// 				ch <- struct{}{}
// 				return nil
// 			}),
// 		}).Run(ctx))
// 	}
//
// 	wg.Wait()
// 	assert.Len(t, ch, 200)
// 	assert.Empty(t, buffer.String())
// }
