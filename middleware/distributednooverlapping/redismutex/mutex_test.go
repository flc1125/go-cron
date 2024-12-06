package redismutex

import (
	"context"
	"testing"
	"time"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/v4/middleware/distributednooverlapping"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
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
		client.Close()
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

		acquired, err := mutex.Lock(ctx, job)
		assert.NoError(t, err)
		assert.True(t, acquired)

		acquired, err = mutex.Lock(ctx, job)
		assert.NoError(t, err)
		assert.False(t, acquired)

		time.Sleep(time.Second + time.Millisecond*10)

		acquired, err = mutex.Lock(ctx, job)
		assert.NoError(t, err)
		assert.True(t, acquired)

		// unlock
		err = mutex.Unlock(ctx, job)
		assert.NoError(t, err)

		acquired, err = mutex.Lock(ctx, job)
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
		acquired, err := mutex.Lock(ctx, job1)
		assert.NoError(t, err)
		assert.True(t, acquired)

		// lock job2
		acquired, err = mutex.Lock(ctx, job2)
		assert.NoError(t, err)
		assert.True(t, acquired)

		acquired, err = mutex.Lock(ctx, job1)
		assert.NoError(t, err)
		assert.False(t, acquired)

		acquired, err = mutex.Lock(ctx, job2)
		assert.NoError(t, err)
		assert.False(t, acquired)

		// unlock job1
		err = mutex.Unlock(ctx, job1)
		assert.NoError(t, err)

		acquired, err = mutex.Lock(ctx, job1)
		assert.NoError(t, err)
		assert.True(t, acquired)

		// unlock job2
		err = mutex.Unlock(ctx, job1)
		assert.NoError(t, err)

		// unlock
		err = mutex.Unlock(ctx, job2)
		assert.NoError(t, err)

		acquired, err = mutex.Lock(ctx, job2)
		assert.NoError(t, err)
		assert.True(t, acquired)
	})
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
