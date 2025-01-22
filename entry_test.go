package cron

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntry_Attributes(t *testing.T) {
	entry := NewEntry(1, nil, JobFunc(func(context.Context) error {
		return nil
	}))
	assert.Equal(t, entry.ID(), EntryID(1))
	assert.NotNil(t, entry.WrappedJob())
	assert.NotNil(t, entry.Job())
	assert.Nil(t, entry.Schedule())
	assert.Zero(t, entry.Next())
	assert.Zero(t, entry.Prev())
	assert.True(t, entry.Valid())
}

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
			entry = NewEntry(tt.id, nil, JobFunc(func(ctx context.Context) error {
				entry, ok := EntryFromContext(ctx)
				assert.True(t, ok)
				assert.Equal(t, entry.ID(), tt.id)

				return nil
			}))

			assert.NoError(t, entry.WrappedJob().Run(ctx))
		})
	}
}

func TestEntry_ContextUseCron(t *testing.T) {
	cron := newWithSeconds()
	var e1, e2 atomic.Value
	var wg sync.WaitGroup
	wg.Add(2)
	_, err := cron.AddFunc("* * * * *", func(ctx context.Context) error {
		defer wg.Done()
		entry, ok := EntryFromContext(ctx)
		assert.True(t, ok)
		assert.True(t, entry.Valid())
		e1.Store(entry)

		t.Logf("entry id: %d", entry.ID())

		return nil
	})
	assert.NoError(t, err)

	_, err = cron.AddFunc("* * * * *", func(ctx context.Context) error {
		defer wg.Done()
		entry, ok := EntryFromContext(ctx)
		assert.True(t, ok)
		assert.True(t, entry.Valid())
		e2.Store(entry)

		t.Logf("entry id: %d", entry.ID())

		return nil
	})
	assert.NoError(t, err)

	cron.Start()
	defer cron.Stop()

	// wait for the job to run
	wg.Wait()

	// ensure the entries are different
	assert.NotNil(t, e1.Load())
	assert.NotNil(t, e2.Load())
	assert.NotEqual(t, e1.Load().(*Entry).id, e2.Load().(*Entry).id)
}
