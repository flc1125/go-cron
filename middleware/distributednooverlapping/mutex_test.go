package distributednooverlapping

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMutex_NoopMutex(t *testing.T) {
	m := NoopMutex{}
	ctx := t.Context()

	lock, acquired, err := m.Lock(ctx, nil)
	assert.NoError(t, err)
	assert.True(t, acquired)
	assert.NotNil(t, lock)

	err = lock.Unlock(ctx)
	assert.NoError(t, err)
}
