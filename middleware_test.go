package cron

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func appendingJob(slice *[]int, value int) Job {
	var m sync.Mutex
	return JobFunc(func(context.Context) error {
		m.Lock()
		defer m.Unlock()
		*slice = append(*slice, value)
		return nil
	})
}

func appendingWrapper(slice *[]int, value int) Middleware {
	return func(j Job) Job {
		return JobFunc(func(ctx context.Context) error {
			appendingJob(slice, value).Run(ctx) //nolint:errcheck
			return j.Run(ctx)
		})
	}
}

func TestChain(t *testing.T) {
	var nums []int
	var (
		append1 = appendingWrapper(&nums, 1)
		append2 = appendingWrapper(&nums, 2)
		append3 = appendingWrapper(&nums, 3)
		append4 = appendingJob(&nums, 4)
	)
	Chain(append1, append2, append3)(append4).Run(context.Background()) //nolint:errcheck
	if !reflect.DeepEqual(nums, []int{1, 2, 3, 4}) {
		t.Error("unexpected order of calls:", nums)
	}
}

type countJob struct {
	m       sync.Mutex
	started int
	done    int
	delay   time.Duration
}

func (j *countJob) Run(context.Context) error {
	j.m.Lock()
	j.started++
	j.m.Unlock()
	time.Sleep(j.delay)
	j.m.Lock()
	j.done++
	j.m.Unlock()
	return nil
}

func (j *countJob) Started() int {
	defer j.m.Unlock()
	j.m.Lock()
	return j.started
}

func (j *countJob) Done() int {
	defer j.m.Unlock()
	j.m.Lock()
	return j.done
}

func TestChainDelayIfStillRunning(t *testing.T) {
	t.Run("runs immediately", func(t *testing.T) {
		var j countJob
		wrappedJob := Chain(DelayIfStillRunning(DiscardLogger))(&j)
		go wrappedJob.Run(context.Background()) //nolint:errcheck
		time.Sleep(2 * time.Millisecond)        // Give the job 2ms to complete.
		if c := j.Done(); c != 1 {
			t.Errorf("expected job run once, immediately, got %d", c)
		}
	})

	t.Run("second run immediate if first done", func(t *testing.T) {
		var j countJob
		wrappedJob := Chain(DelayIfStillRunning(DiscardLogger))(&j)
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			go func() {
				defer wg.Done()
				wrappedJob.Run(context.Background()) //nolint:errcheck
			}()
			time.Sleep(time.Millisecond)
			go func() {
				defer wg.Done()
				wrappedJob.Run(context.Background()) //nolint:errcheck
			}()
		}()
		wg.Wait()
		if c := j.Done(); c != 2 {
			t.Errorf("expected job run twice, immediately, got %d", c)
		}
	})

	t.Run("second run delayed if first not done", func(t *testing.T) {
		var j countJob
		j.delay = 100 * time.Millisecond
		wrappedJob := Chain(DelayIfStillRunning(DiscardLogger))(&j)
		go func() {
			go wrappedJob.Run(context.Background()) //nolint:errcheck
			time.Sleep(10 * time.Millisecond)
			go wrappedJob.Run(context.Background()) //nolint:errcheck
		}()

		// After 50ms, the first job is still in progress, and the second job was
		// run but should be waiting for it to finish.
		time.Sleep(50 * time.Millisecond)
		started, done := j.Started(), j.Done()
		if started != 1 || done != 0 {
			t.Error("expected first job started, but not finished, got", started, done)
		}

		// Verify that the second job completes.
		time.Sleep(250 * time.Millisecond)
		started, done = j.Started(), j.Done()
		if started != 2 || done != 2 {
			t.Error("expected both jobs done, got", started, done)
		}
	})
}

func TestMiddleware_NoopMiddleware(t *testing.T) {
	err := NoopMiddleware()(JobFunc(func(context.Context) error {
		return nil
	})).Run(context.Background())
	assert.NoError(t, err)
}
