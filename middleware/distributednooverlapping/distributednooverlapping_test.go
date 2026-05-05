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

type spyMutex struct {
	lock     Lock
	acquired bool
	err      error

	lockCalls int
	job       JobWithMutex
}

func (m *spyMutex) Lock(_ context.Context, job JobWithMutex) (Lock, bool, error) {
	m.lockCalls++
	m.job = job
	return m.lock, m.acquired, m.err
}

type spyLock struct {
	err         error
	unlockCalls int
}

func (l *spyLock) Unlock(context.Context) error {
	l.unlockCalls++
	return l.err
}

type logCall struct {
	err error
	msg string
}

type spyLogger struct {
	infos  []logCall
	errors []logCall
}

func (l *spyLogger) Info(msg string, _ ...any) {
	l.infos = append(l.infos, logCall{msg: msg})
}

func (l *spyLogger) Error(err error, msg string, _ ...any) {
	l.errors = append(l.errors, logCall{err: err, msg: msg})
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

func TestMiddleware_NoEntryContextRunsOriginalJob(t *testing.T) {
	mutex := &spyMutex{}
	middleware := New(mutex)
	ran := false

	err := middleware(cron.JobFunc(func(context.Context) error {
		ran = true
		return nil
	})).Run(ctx)

	assert.NoError(t, err)
	assert.True(t, ran)
	assert.Zero(t, mutex.lockCalls)
}

func TestMiddleware_NonMutexEntryJobRunsOriginalJob(t *testing.T) {
	mutex := &spyMutex{}
	middleware := New(mutex)
	entry := entryForJob(cron.NoopJob{})
	ran := false

	err := middleware(cron.JobFunc(func(context.Context) error {
		ran = true
		return nil
	})).Run(cron.WithEntryContext(ctx, &entry))

	assert.NoError(t, err)
	assert.True(t, ran)
	assert.Zero(t, mutex.lockCalls)
}

func TestMiddleware_LockErrorReturnsError(t *testing.T) {
	lockErr := errors.New("lock failed")
	mutex := &spyMutex{err: lockErr}
	logger := &spyLogger{}
	middleware := New(mutex, WithLogger(logger))
	entry := entryForJob(testJob{
		name: "test",
		ttl:  time.Second,
		Job:  cron.NoopJob{},
	})
	ran := false

	err := middleware(cron.JobFunc(func(context.Context) error {
		ran = true
		return nil
	})).Run(cron.WithEntryContext(ctx, &entry))

	assert.ErrorIs(t, err, lockErr)
	assert.False(t, ran)
	assert.Equal(t, 1, mutex.lockCalls)
	assert.Equal(t, "test", mutex.job.GetMutexKey())
	assert.Len(t, logger.errors, 1)
	assert.ErrorIs(t, logger.errors[0].err, lockErr)
}

func TestMiddleware_UnacquiredLockSkipsJob(t *testing.T) {
	mutex := &spyMutex{acquired: false}
	logger := &spyLogger{}
	middleware := New(mutex, WithLogger(logger))
	entry := entryForJob(testJob{
		name: "test",
		ttl:  time.Second,
		Job:  cron.NoopJob{},
	})
	ran := false

	err := middleware(cron.JobFunc(func(context.Context) error {
		ran = true
		return nil
	})).Run(cron.WithEntryContext(ctx, &entry))

	assert.NoError(t, err)
	assert.False(t, ran)
	assert.Equal(t, 1, mutex.lockCalls)
	assert.Len(t, logger.infos, 1)
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

func TestMiddleware_SuccessfulJobUnlocks(t *testing.T) {
	lock := &spyLock{}
	mutex := &spyMutex{lock: lock, acquired: true}
	middleware := New(mutex)
	entry := entryForJob(testJob{
		name: "test",
		ttl:  time.Second,
		Job:  cron.NoopJob{},
	})
	ran := false

	err := middleware(cron.JobFunc(func(context.Context) error {
		ran = true
		return nil
	})).Run(cron.WithEntryContext(ctx, &entry))

	assert.NoError(t, err)
	assert.True(t, ran)
	assert.Equal(t, 1, mutex.lockCalls)
	assert.Equal(t, 1, lock.unlockCalls)
}

func TestMiddleware_UnlockErrorDoesNotOverrideJobError(t *testing.T) {
	jobErr := errors.New("job failed")
	unlockErr := errors.New("unlock failed")
	lock := &spyLock{err: unlockErr}
	mutex := &spyMutex{lock: lock, acquired: true}
	logger := &spyLogger{}
	middleware := New(mutex, WithLogger(logger))
	entry := entryForJob(testJob{
		name: "test",
		ttl:  time.Second,
		Job:  cron.NoopJob{},
	})

	err := middleware(cron.JobFunc(func(context.Context) error {
		return jobErr
	})).Run(cron.WithEntryContext(ctx, &entry))

	assert.ErrorIs(t, err, jobErr)
	assert.Equal(t, 1, lock.unlockCalls)
	assert.Len(t, logger.errors, 1)
	assert.ErrorIs(t, logger.errors[0].err, unlockErr)
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
