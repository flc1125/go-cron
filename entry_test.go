package cron

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntry_Attributes(t *testing.T) {
	entry := newEntry(1, nil, JobFunc(func(context.Context) error {
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
			ctx := t.Context()

			// non-existent entry
			entry, ok := EntryFromContext(ctx)
			assert.False(t, ok)
			assert.Nil(t, entry)

			// existent entry
			entry = newEntry(tt.id, nil, JobFunc(func(ctx context.Context) error {
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

func TestEntry_ExecutionContextIsScopedToRegisteredEntry(t *testing.T) {
	var firstSeen, secondSeen *Entry
	first := newEntry(1, nil, JobFunc(func(ctx context.Context) error {
		firstSeen, _ = EntryFromContext(ctx)
		return nil
	}))
	second := newEntry(2, nil, JobFunc(func(ctx context.Context) error {
		secondSeen, _ = EntryFromContext(ctx)
		return nil
	}))
	executionSnapshot := *first
	executionSnapshot.prev = time.Unix(1, 0)
	executionSnapshot.next = time.Unix(2, 0)
	executionCtx := withExecutionEntryContext(t.Context(), first, &executionSnapshot)

	assert.NoError(t, first.WrappedJob().Run(executionCtx))
	assert.Same(t, &executionSnapshot, firstSeen)

	assert.NoError(t, second.WrappedJob().Run(executionCtx))
	assert.Same(t, second, secondSeen)

	assert.NoError(t, first.WrappedJob().Run(t.Context()))
	assert.Same(t, first, firstSeen)
}

type entrySnapshotSchedule time.Duration

func (s entrySnapshotSchedule) Next(now time.Time) time.Time {
	return now.Add(time.Duration(s))
}

func TestCron_EntryExecutionSnapshot(t *testing.T) {
	entries := make(chan *Entry, 16)
	release := make(chan struct{})
	job := JobFunc(func(ctx context.Context) error {
		entry, _ := EntryFromContext(ctx)
		entries <- entry
		for {
			select {
			case <-release:
				return nil
			default:
				_ = entry.Prev()
				_ = entry.Next()
				runtime.Gosched()
			}
		}
	})

	c := New()
	id := c.Schedule(entrySnapshotSchedule(20*time.Millisecond), job)
	c.Start()
	defer c.Stop()

	readEntry := func() *Entry {
		t.Helper()
		select {
		case entry := <-entries:
			require.NotNil(t, entry)
			return entry
		case <-time.After(2 * time.Second):
			require.FailNow(t, "timed out waiting for scheduled entry")
			return nil
		}
	}

	first := readEntry()
	firstPrev, firstNext := first.Prev(), first.Next()
	second := readEntry()
	close(release)
	<-c.Stop().Done()

	assert.Equal(t, id, first.ID())
	assert.Equal(t, id, second.ID())
	assert.NotSame(t, first, second)
	assert.False(t, firstPrev.IsZero())
	assert.True(t, firstNext.After(firstPrev))
	assert.Equal(t, firstNext, second.Prev())
	assert.True(t, second.Next().After(second.Prev()))
	assert.Equal(t, firstPrev, first.Prev())
	assert.Equal(t, firstNext, first.Next())
}
