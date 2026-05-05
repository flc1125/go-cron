package redismutex

import (
	"context"
	"testing"
	"time"

	"github.com/flc1125/go-cron/middleware/distributednooverlapping/v4"
	"github.com/flc1125/go-cron/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

type testJob struct {
	t    *testing.T
	job  cron.JobFunc
	name string
	ttl  time.Duration
}

var _ distributednooverlapping.JobWithMutex = (*testJob)(nil)

func (t testJob) Run(ctx context.Context) error {
	return t.job(ctx)
}

func (t testJob) GetMutexKey() string {
	return t.name
}

func (t testJob) GetMutexTTL() time.Duration {
	return t.ttl
}

func createRedis(t *testing.T) redis.UniversalClient {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	t.Cleanup(func() {
		client.FlushAll(ctx)
		client.Close() //nolint: errcheck
	})
	return client
}

func TestMutex(t *testing.T) {
	client := createRedis(t)
	mutex := New(client, WithPrefix("test:cron"))

	t.Run("basic routine testing", func(t *testing.T) {
		job := testJob{
			t: t,
			job: func(context.Context) error {
				return nil
			},
			name: "job:basic",
			ttl:  time.Second,
		}

		lock, acquired, err := mutex.Lock(ctx, job)
		assert.NoError(t, err)
		assert.True(t, acquired)
		assert.NotNil(t, lock)

		_, acquired, err = mutex.Lock(ctx, job)
		assert.NoError(t, err)
		assert.False(t, acquired)

		// unlock
		err = lock.Unlock(ctx)
		assert.NoError(t, err)

		lock, acquired, err = mutex.Lock(ctx, job)
		assert.NoError(t, err)
		assert.True(t, acquired)
		assert.NotNil(t, lock)

		// unlock
		err = lock.Unlock(ctx)
		assert.NoError(t, err)

		_, acquired, err = mutex.Lock(ctx, job)
		assert.NoError(t, err)
		assert.True(t, acquired)
	})

	t.Run("multiple jobs to see if there is mutual exclusion", func(t *testing.T) {
		job1 := testJob{
			t: t,
			job: func(context.Context) error {
				return nil
			},
			name: "job:multi1",
			ttl:  time.Second,
		}

		job2 := testJob{
			t: t,
			job: func(context.Context) error {
				return nil
			},
			name: "job:multi2",
			ttl:  time.Second,
		}

		// lock job1
		lock1, acquired, err := mutex.Lock(ctx, job1)
		assert.NoError(t, err)
		assert.True(t, acquired)
		assert.NotNil(t, lock1)

		// lock job2
		lock2, acquired, err := mutex.Lock(ctx, job2)
		assert.NoError(t, err)
		assert.True(t, acquired)
		assert.NotNil(t, lock2)

		_, acquired, err = mutex.Lock(ctx, job1)
		assert.NoError(t, err)
		assert.False(t, acquired)

		_, acquired, err = mutex.Lock(ctx, job2)
		assert.NoError(t, err)
		assert.False(t, acquired)

		// unlock job1
		err = lock1.Unlock(ctx)
		assert.NoError(t, err)

		lock1, acquired, err = mutex.Lock(ctx, job1)
		assert.NoError(t, err)
		assert.True(t, acquired)
		assert.NotNil(t, lock1)

		// unlock job1
		err = lock1.Unlock(ctx)
		assert.NoError(t, err)

		// unlock job2
		err = lock2.Unlock(ctx)
		assert.NoError(t, err)

		_, acquired, err = mutex.Lock(ctx, job2)
		assert.NoError(t, err)
		assert.True(t, acquired)
	})
}

func TestMutex_UnlockDoesNotDeleteSuccessorLock(t *testing.T) {
	client := createRedis(t)
	staleOwner := New(client, WithPrefix("test:cron"))
	successor := New(client, WithPrefix("test:cron"))
	contender := New(client, WithPrefix("test:cron"))

	staleJob := testJob{
		t: t,
		job: func(context.Context) error {
			return nil
		},
		name: "job:stale-owner",
		ttl:  50 * time.Millisecond,
	}
	successorJob := testJob{
		t: t,
		job: func(context.Context) error {
			return nil
		},
		name: staleJob.name,
		ttl:  time.Second,
	}

	staleLock, acquired, err := staleOwner.Lock(ctx, staleJob)
	require.NoError(t, err)
	require.True(t, acquired)
	require.NotNil(t, staleLock)

	var successorLock distributednooverlapping.Lock
	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		var err error
		successorLock, acquired, err = successor.Lock(ctx, successorJob)
		assert.NoError(collect, err)
		assert.True(collect, acquired)
	}, time.Second, 10*time.Millisecond)
	require.NotNil(t, successorLock)

	err = staleLock.Unlock(ctx)
	require.NoError(t, err)

	_, acquired, err = contender.Lock(ctx, successorJob)
	require.NoError(t, err)
	assert.False(t, acquired, "stale unlock must not remove the successor lock")

	err = successorLock.Unlock(ctx)
	require.NoError(t, err)

	_, acquired, err = contender.Lock(ctx, successorJob)
	require.NoError(t, err)
	assert.True(t, acquired)
}

func TestMutex_Prefix(t *testing.T) {
	// without prefix
	t.Run("without prefix", func(t *testing.T) {
		assert.Equal(t, "cron:", New(nil).prefix)
	})

	// with prefix
	t.Run("with prefix", func(t *testing.T) {
		tests := []struct {
			name   string
			prefix string
			want   string
		}{
			{"", "test", "test:"},
			{"", "test:", "test:"},
			{"", "", "cron:"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.want, New(nil, WithPrefix(tt.prefix)).prefix)
			})
		}
	})
}
