package redismutex // import "github.com/flc1125/go-cron/v4/middleware/distributednooverlapping/redismutex"

import (
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/flc1125/go-cron/v4/middleware/distributednooverlapping"
)

type Mutex struct {
	redis  redis.UniversalClient
	prefix string
}

type Option func(*Mutex)

func WithPrefix(prefix string) Option {
	return func(m *Mutex) {
		if prefix != "" {
			if prefix[len(prefix)-1] == ':' {
				prefix = prefix[:len(prefix)-1]
			}
			m.prefix = prefix + ":"
		}
	}
}

var _ distributednooverlapping.Mutex = (*Mutex)(nil)

func New(redis redis.UniversalClient, opts ...Option) *Mutex {
	mutex := &Mutex{
		redis:  redis,
		prefix: "cron:mutex",
	}
	for _, opt := range opts {
		opt(mutex)
	}
	return mutex
}

func (m *Mutex) Lock(ctx context.Context, job distributednooverlapping.JobWithMutex) (bool, error) {
	return m.redis.SetNX(ctx, m.prefix+job.GetMutexKey(), 1, job.GetMutexTTL()).Result()
}

func (m *Mutex) Unlock(ctx context.Context, job distributednooverlapping.JobWithMutex) error {
	return m.redis.Del(ctx, m.prefix+job.GetMutexKey()).Err()
}
