package redismutex

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/flc1125/go-cron/middleware/distributednooverlapping/v4"
	redis "github.com/redis/go-redis/v9"
)

type Mutex struct {
	redis  redis.UniversalClient
	prefix string
}

var _ distributednooverlapping.Mutex = (*Mutex)(nil)

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

var unlockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`)

func New(redis redis.UniversalClient, opts ...Option) *Mutex {
	mutex := &Mutex{
		redis:  redis,
		prefix: "cron:",
	}
	for _, opt := range opts {
		opt(mutex)
	}
	return mutex
}

func (m *Mutex) Lock(ctx context.Context, job distributednooverlapping.JobWithMutex) (distributednooverlapping.Lock, bool, error) {
	key := m.key(job)
	token, err := newToken()
	if err != nil {
		return nil, false, err
	}

	acquired, err := m.redis.SetNX(ctx, key, token, job.GetMutexTTL()).Result()
	if err != nil || !acquired {
		return nil, acquired, err
	}

	return &lock{
		redis: m.redis,
		key:   key,
		token: token,
	}, true, nil
}

type lock struct {
	redis redis.UniversalClient
	key   string
	token string
}

func (l *lock) Unlock(ctx context.Context) error {
	return unlockScript.Run(ctx, l.redis, []string{l.key}, l.token).Err()
}

func newToken() (string, error) {
	var token [32]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", fmt.Errorf("generate redis mutex token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(token[:]), nil
}

func (m *Mutex) key(job distributednooverlapping.JobWithMutex) string {
	return m.prefix + job.GetMutexKey()
}
