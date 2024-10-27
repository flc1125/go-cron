package cron

import (
	"context"
	"testing"
	"time"

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

func TestEntry_ContextUseCron(t *testing.T) {
	cron := newWithSeconds()
	var e1, e2 *Entry
	_, err := cron.AddFunc("* * * * *", func(ctx context.Context) error {
		entry, ok := EntryFromContext(ctx)
		assert.True(t, ok)
		assert.True(t, entry.Valid())
		e1 = entry

		t.Logf("entry id: %d", entry.ID)

		return nil
	})
	assert.NoError(t, err)

	_, err = cron.AddFunc("* * * * *", func(ctx context.Context) error {
		entry, ok := EntryFromContext(ctx)
		assert.True(t, ok)
		assert.True(t, entry.Valid())
		e2 = entry

		t.Logf("entry id: %d", entry.ID)

		return nil
	})
	assert.NoError(t, err)

	cron.Start()
	defer cron.Stop()

	// wait for the job to run
	time.Sleep(time.Second)

	// ensure the entries are different
	assert.NotEqual(t, e1.ID, e2.ID)
}
