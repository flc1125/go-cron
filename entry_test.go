package cron

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntry_Context(t *testing.T) {
	tests := []struct {
		name string
		id   EntryID
	}{
		{"", 1},
		{"", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// non-existent entry
			entry, ok := EntryFromContext(ctx)
			assert.False(t, ok)
			assert.Nil(t, entry)

			// existent entry
			entry = newEntry(tt.id, nil, JobFunc(func(ctx context.Context) error {
				entry, ok := EntryFromContext(ctx)
				assert.True(t, ok)
				assert.Equal(t, entry.ID, tt.id)

				return nil
			}))

			assert.NoError(t, entry.WrappedJob().Run(ctx))
		})
	}
}
