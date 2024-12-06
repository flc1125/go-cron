package distributednooverlapping

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMutex_NoopMutex(t *testing.T) {
	m := NoopMutex{}
	ctx := context.Background()

	acquired, err := m.Lock(ctx, nil)
	assert.NoError(t, err)
	assert.True(t, acquired)

	err = m.Unlock(ctx, nil)
	assert.NoError(t, err)
}
