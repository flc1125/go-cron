package nooverlapping

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flc1125/go-cron/crontest/v4/logger"
	"github.com/flc1125/go-cron/v4"
)

var (
	ctx           = context.Background()
	buf           = logger.NewBuffer()
	noOverlapping = New(WithLogger(logger.NewBufferLogger(buf)))
	wg            = sync.WaitGroup{}
)

func TestNoOverlapping(t *testing.T) {
	buf.Reset()

	var (
		ch  = make(chan struct{}, 100)
		wg  = sync.WaitGroup{}
		job = noOverlapping(cron.JobFunc(func(context.Context) error {
			ch <- struct{}{}
			time.Sleep(2 * time.Millisecond)
			return nil
		}))
	)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.NoError(t, job.Run(ctx))
		}()
	}
	wg.Wait()

	assert.True(t, len(ch) < 100)
	assert.Contains(t, buf.String(), "job is still running, skip")
}

func TestNoOverlapping_Chain(t *testing.T) {
	buf.Reset()

	var (
		ch  = make(chan struct{}, 100)
		job = cron.Chain(noOverlapping)(cron.JobFunc(func(context.Context) error {
			ch <- struct{}{}
			time.Sleep(2 * time.Millisecond)
			return nil
		}))
	)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.NoError(t, job.Run(ctx))
		}()
	}
	wg.Wait()

	assert.True(t, len(ch) < 100)
	assert.Contains(t, buf.String(), "job is still running, skip")
}

func TestNoOverlapping_Cases(t *testing.T) {
	t.Run("second run immediate if first done", func(t *testing.T) {
		buf.Reset()
		ch := make(chan struct{}, 10)
		job := noOverlapping(cron.JobFunc(func(context.Context) error {
			ch <- struct{}{}
			time.Sleep(20 * time.Millisecond)
			return nil
		}))

		wg.Add(3)
		go func() {
			defer wg.Done()
			go func() {
				defer wg.Done()
				assert.NoError(t, job.Run(ctx))
			}()

			time.Sleep(100 * time.Millisecond)

			go func() {
				defer wg.Done()
				assert.NoError(t, job.Run(ctx))
			}()
		}()
		wg.Wait()

		assert.Len(t, ch, 2)
	})

	t.Run("second run skipped if first not done", func(t *testing.T) {
		buf.Reset()
		ch := make(chan struct{}, 10)
		job := noOverlapping(cron.JobFunc(func(context.Context) error {
			ch <- struct{}{}
			time.Sleep(10 * time.Millisecond)
			return nil
		}))

		wg.Add(3)
		go func() {
			defer wg.Done()
			go func() {
				defer wg.Done()
				assert.NoError(t, job.Run(ctx))
			}()

			go func() {
				defer wg.Done()
				assert.NoError(t, job.Run(ctx))
			}()
		}()
		wg.Wait()

		assert.Len(t, ch, 1)
		assert.Contains(t, buf.String(), "job is still running, skip")
	})

	t.Run("skip 10 jobs on rapid fire", func(t *testing.T) {
		buf.Reset()
		ch := make(chan struct{}, 10)
		job := noOverlapping(cron.JobFunc(func(context.Context) error {
			ch <- struct{}{}
			time.Sleep(100 * time.Millisecond)
			return nil
		}))

		wg.Add(10)
		for i := 0; i < 10; i++ {
			go func() {
				defer wg.Done()
				assert.NoError(t, job.Run(ctx))
			}()
		}
		wg.Wait()

		assert.Len(t, ch, 1)
		assert.Contains(t, buf.String(), "job is still running, skip")
	})

	t.Run("different jobs independent", func(t *testing.T) {
		buf.Reset()
		ch := make(chan struct{}, 100)
		job1 := noOverlapping(cron.JobFunc(func(context.Context) error {
			ch <- struct{}{}
			time.Sleep(100 * time.Millisecond)
			return nil
		}))
		job2 := noOverlapping(cron.JobFunc(func(context.Context) error {
			ch <- struct{}{}
			time.Sleep(100 * time.Millisecond)
			return nil
		}))

		for i := 0; i < 10; i++ {
			wg.Add(2)
			go func() {
				defer wg.Done()
				assert.NoError(t, job1.Run(ctx))
			}()
			go func() {
				defer wg.Done()
				assert.NoError(t, job2.Run(ctx))
			}()
		}

		wg.Wait()
		assert.Len(t, ch, 2)
		assert.Contains(t, buf.String(), "job is still running, skip")
	})
}
