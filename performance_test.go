package cron

import (
	"context"
	"testing"
	"time"
)

func TestSchedulerPerformanceWithManyEntries(t *testing.T) {
	c := New()

	jobCount := 100
	for i := 0; i < jobCount; i++ {
		_, err := c.AddFunc("@every 1h", func(ctx context.Context) error {
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to add job %d: %v", i, err)
		}
	}

	if len(c.entries) != jobCount {
		t.Errorf("Expected %d entries, got %d", jobCount, len(c.entries))
	}

	c.Start()
	defer c.Stop()

	time.Sleep(10 * time.Millisecond)

	entries := c.Entries()
	if len(entries) != jobCount {
		t.Errorf("Expected %d entries after start, got %d", jobCount, len(entries))
	}

	for i, entry := range entries {
		if entry.Next().IsZero() {
			t.Errorf("Entry %d has zero next time", i)
		}
	}
}

func TestConditionalSortingBehavior(t *testing.T) {
	c := New()

	if c.needsSort {
		t.Error("needsSort should be false initially")
	}

	_, err := c.AddFunc("@every 1h", func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	if !c.needsSort {
		t.Error("needsSort should be true after adding entry")
	}

	c.Start()
	defer c.Stop()

	time.Sleep(10 * time.Millisecond)

	if c.needsSort {
		t.Error("needsSort should be false after scheduler processes entries")
	}
}
