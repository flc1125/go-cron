package distributednooverlapping

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMutex_NoopMutex(t *testing.T) {
	m := NoopMutex{}

	acquired, err := m.Lock(nil, nil)
	assert.NoError(t, err)
	assert.True(t, acquired)

	err = m.Unlock(nil, nil)
	assert.NoError(t, err)
}
