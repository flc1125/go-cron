package delayoverlapping

import (
	"bytes"
	"context"
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
	wg          = sync.WaitGroup{}
)

func newBufferLogger() (*bytes.Buffer, cron.Logger) {
	buf := new(bytes.Buffer)
	return buf, cron.VerbosePrintfLogger(log.New(buf, "", log.LstdFlags))
}

func TestDelayOverlapping(t *testing.T) {
	buf.Reset()

	var (
		delayOverlapping = New(
			WithLogger(logger),
			WithReminderTime(1*time.Millisecond),
		)
		ch  = make(chan struct{}, 100)
		job = delayOverlapping(cron.JobFunc(func(context.Context) error {
			ch <- struct{}{}
			time.Sleep(2 * time.Millisecond)
			return nil
		}))
	)

	starting := time.Now()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.NoError(t, job.Run(ctx))
		}()
	}

	wg.Wait()

	assert.True(t, len(ch) == 100)
	assert.Contains(t, buf.String(), "delay")
	assert.Greater(t, time.Since(starting).Milliseconds(), int64(100))
}

func TestDelayOverlapping_Chain(t *testing.T) {
	buf.Reset()

	var (
		delayOverlapping = New(
			WithLogger(logger),
			WithReminderTime(1*time.Millisecond),
		)
		ch  = make(chan struct{}, 100)
		job = cron.Chain(delayOverlapping)(cron.JobFunc(func(context.Context) error {
			ch <- struct{}{}
			time.Sleep(2 * time.Millisecond)
			return nil
		}))
	)

	starting := time.Now()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.NoError(t, job.Run(ctx))
		}()
	}

	wg.Wait()

	assert.True(t, len(ch) == 100)
	assert.Contains(t, buf.String(), "delay")
	assert.Greater(t, time.Since(starting).Milliseconds(), int64(100))
}
