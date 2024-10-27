package nooverlapping

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

func TestNoOverlapping(t *testing.T) {
	var (
		buf = logger.NewBuffer()
		m   = New(WithLogger(logger.NewBufferLogger(buf)))
		ch  = make(chan struct{}, 100)
		wg  = sync.WaitGroup{}
		job = m(cron.JobFunc(func(context.Context) error {
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

	assert.Len(t, ch, 1)
	assert.Contains(t, buf.String(), "job is still running, skip")
}

// func TestChainSkipIfStillRunning(t *testing.T) {
// 	t.Run("runs immediately", func(t *testing.T) {
// 		var j countJob
// 		wrappedJob := Chain(SkipIfStillRunning(DiscardLogger))(&j)
// 		go wrappedJob.Run(context.Background()) //nolint:errcheck
// 		time.Sleep(20 * time.Millisecond)       // Give the job 2ms to complete.
// 		if c := j.Done(); c != 1 {
// 			t.Errorf("expected job run once, immediately, got %d", c)
// 		}
// 	})
//
// 	t.Run("second run immediate if first done", func(t *testing.T) {
// 		var j countJob
// 		wrappedJob := Chain(SkipIfStillRunning(DiscardLogger))(&j)
// 		go func() {
// 			go wrappedJob.Run(context.Background()) //nolint:errcheck
// 			time.Sleep(10 * time.Millisecond)
// 			go wrappedJob.Run(context.Background()) //nolint:errcheck
// 		}()
// 		time.Sleep(30 * time.Millisecond) // Give both jobs 3ms to complete.
// 		if c := j.Done(); c != 2 {
// 			t.Errorf("expected job run twice, immediately, got %d", c)
// 		}
// 	})
//
// 	t.Run("second run skipped if first not done", func(t *testing.T) {
// 		var j countJob
// 		j.delay = 10 * time.Millisecond
// 		wrappedJob := Chain(SkipIfStillRunning(DiscardLogger))(&j)
// 		go func() {
// 			go wrappedJob.Run(context.Background()) //nolint:errcheck
// 			time.Sleep(time.Millisecond)
// 			go wrappedJob.Run(context.Background()) //nolint:errcheck
// 		}()
//
// 		// After 5ms, the first job is still in progress, and the second job was
// 		// aleady skipped.
// 		time.Sleep(5 * time.Millisecond)
// 		started, done := j.Started(), j.Done()
// 		if started != 1 || done != 0 {
// 			t.Error("expected first job started, but not finished, got", started, done)
// 		}
//
// 		// Verify that the first job completes and second does not run.
// 		time.Sleep(25 * time.Millisecond)
// 		started, done = j.Started(), j.Done()
// 		if started != 1 || done != 1 {
// 			t.Error("expected second job skipped, got", started, done)
// 		}
// 	})
//
// 	t.Run("skip 10 jobs on rapid fire", func(t *testing.T) {
// 		var j countJob
// 		j.delay = 10 * time.Millisecond
// 		wrappedJob := Chain(SkipIfStillRunning(DiscardLogger))(&j)
// 		for i := 0; i < 11; i++ {
// 			go wrappedJob.Run(context.Background()) //nolint:errcheck
// 		}
// 		time.Sleep(200 * time.Millisecond)
// 		done := j.Done()
// 		if done != 1 {
// 			t.Error("expected 1 jobs executed, 10 jobs dropped, got", done)
// 		}
// 	})
//
// 	t.Run("different jobs independent", func(t *testing.T) {
// 		var j1, j2 countJob
// 		j1.delay = 10 * time.Millisecond
// 		j2.delay = 10 * time.Millisecond
// 		chain := Chain(SkipIfStillRunning(DiscardLogger))
// 		wrappedJob1 := chain(&j1)
// 		wrappedJob2 := chain(&j2)
// 		for i := 0; i < 11; i++ {
// 			go wrappedJob1.Run(context.Background()) //nolint:errcheck
// 			go wrappedJob2.Run(context.Background()) //nolint:errcheck
// 		}
// 		time.Sleep(100 * time.Millisecond)
// 		var (
// 			done1 = j1.Done()
// 			done2 = j2.Done()
// 		)
// 		if done1 != 1 || done2 != 1 {
// 			t.Error("expected both jobs executed once, got", done1, "and", done2)
// 		}
// 	})
// }
