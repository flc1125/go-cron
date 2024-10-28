package cron

import (
	"context"
	"reflect"
	"sync"
	"testing"

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

func TestMiddleware_NoopMiddleware(t *testing.T) {
	err := NoopMiddleware()(JobFunc(func(context.Context) error {
		return nil
	})).Run(context.Background())
	assert.NoError(t, err)
}
