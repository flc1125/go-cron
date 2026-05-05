package distributednooverlapping

import (
	"bytes"
	"context"
	"errors"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/flc1125/go-cron/v4"
	"github.com/stretchr/testify/assert"
)

var (
	ctx         = context.Background()
	buf, logger = newBufferLogger()
)

func newBufferLogger() (*bytes.Buffer, cron.Logger) {
	buf := new(bytes.Buffer)
	return buf, cron.VerbosePrintfLogger(log.New(buf, "", log.LstdFlags))
}

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
	t *testing.T //nolint:unused
}

var _ Mutex = testMutex{}

func (m testMutex) Lock(_ context.Context, job JobWithMutex) (Lock, bool, error) {
	if job.GetMutexKey() == "test" {
		return noopLock{}, true, nil
	}

	return nil, false, nil
}

type nilLockMutex struct{}

func (nilLockMutex) Lock(context.Context, JobWithMutex) (Lock, bool, error) {
	return nil, true, nil
}

type contextCheckingMutex struct {
	unlockErr chan<- error
}

func (m contextCheckingMutex) Lock(context.Context, JobWithMutex) (Lock, bool, error) {
	return contextCheckingLock(m), true, nil
}

type contextCheckingLock struct {
	unlockErr chan<- error
}

func (l contextCheckingLock) Unlock(ctx context.Context) error {
	l.unlockErr <- ctx.Err()
	return nil
}

func TestMiddleware_Noop(t *testing.T) {
	buf.Reset()

	ch := make(chan struct{}, 200)
	wg := sync.WaitGroup{}

	noopMiddleware := New(
		NoopMutex{},
		WithLogger(logger),
	)

	for range 100 {
		wg.Add(2)

		// not mutex job, so no blocking
		go assert.NoError(t, noopMiddleware(cron.JobFunc(func(context.Context) error {
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
			Job: cron.JobFunc(func(context.Context) error {
				defer wg.Done()
				time.Sleep(1 * time.Millisecond)
				ch <- struct{}{}
				return nil
			}),
		}).Run(ctx))
	}

	wg.Wait()
	assert.Len(t, ch, 200)
	assert.Empty(t, buf.String())
}

func TestMiddleware_NilAcquiredLock(t *testing.T) {
	middleware := New(nilLockMutex{}, WithLogger(logger))
	entry := entryForJob(testJob{
		name: "test",
		ttl:  time.Second,
		Job: cron.JobFunc(func(context.Context) error {
			return errors.New("job should not run")
		}),
	})

	err := middleware(testJob{
		name: "test",
		ttl:  time.Second,
		Job:  cron.NoopJob{},
	}).Run(cron.WithEntryContext(ctx, &entry))
	assert.EqualError(t, err, "mutex acquired without lock")
}

func TestMiddleware_UnlockWithoutCanceledContext(t *testing.T) {
	unlockErr := make(chan error, 1)
	middleware := New(contextCheckingMutex{unlockErr: unlockErr}, WithLogger(logger))
	entry := entryForJob(testJob{
		name: "test",
		ttl:  time.Second,
		Job:  cron.NoopJob{},
	})
	canceledCtx, cancel := context.WithCancel(ctx)

	err := middleware(testJob{
		name: "test",
		ttl:  time.Second,
		Job: cron.JobFunc(func(context.Context) error {
			cancel()
			return nil
		}),
	}).Run(cron.WithEntryContext(canceledCtx, &entry))
	assert.NoError(t, err)
	assert.NoError(t, <-unlockErr)
}

func entryForJob(job cron.Job) cron.Entry {
	c := cron.New()
	id := c.Schedule(cron.Every(time.Second), job)
	return c.Entry(id)
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
