package cron

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntry_Context(t *testing.T) {
	entry := newEntry(1, nil, JobFunc(func(ctx context.Context) error {
		entry, ok := EntryFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, entry.ID, EntryID(1))

		return nil
	}))

	assert.NoError(t, entry.WrappedJob().Run(context.Background()))
}
