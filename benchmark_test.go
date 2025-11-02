package cron

import (
	"context"
	"testing"
	"time"
)

// BenchmarkParseStandard benchmarks the standard cron parser
func BenchmarkParseStandard(b *testing.B) {
	specs := []string{
		"* * * * *",
		"0 0 * * *",
		"*/5 * * * *",
		"0 0 1 * *",
		"@hourly",
		"@daily",
		"@every 1h",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec := specs[i%len(specs)]
		_, _ = ParseStandard(spec)
	}
}

// BenchmarkScheduleNext benchmarks the SpecSchedule.Next calculation
func BenchmarkScheduleNext(b *testing.B) {
	schedule, _ := ParseStandard("*/5 * * * *")
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = schedule.Next(now)
	}
}

// BenchmarkEvery benchmarks the Every function
func BenchmarkEvery(b *testing.B) {
	durations := []time.Duration{
		time.Second,
		time.Minute,
		time.Hour,
		5 * time.Minute,
		30 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := durations[i%len(durations)]
		_ = Every(d)
	}
}

// BenchmarkAddJob benchmarks adding jobs to the cron scheduler
func BenchmarkAddJob(b *testing.B) {
	job := JobFunc(func(ctx context.Context) error {
		return nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := New()
		_, _ = c.AddJob("* * * * *", job)
	}
}

// BenchmarkStartStop benchmarks starting and stopping the cron scheduler
func BenchmarkStartStop(b *testing.B) {
	c := New()
	_, _ = c.AddFunc("* * * * *", func(ctx context.Context) error {
		return nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Start()
		c.Stop()
	}
}

// BenchmarkJobExecution benchmarks job execution overhead
func BenchmarkJobExecution(b *testing.B) {
	executed := 0
	c := New(WithSeconds())
	_, _ = c.AddFunc("* * * * * *", func(ctx context.Context) error {
		executed++
		return nil
	}, NoopMiddleware())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		job := c.entries[0].WrappedJob()
		_ = job.Run(b.Context())
	}
}

// BenchmarkEntrySnapshot benchmarks creating entry snapshots
func BenchmarkEntrySnapshot(b *testing.B) {
	c := New()
	// Add 10 jobs
	for i := 0; i < 10; i++ {
		_, _ = c.AddFunc("* * * * *", func(ctx context.Context) error {
			return nil
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Entries()
	}
}

// BenchmarkRemoveEntry benchmarks removing entries
func BenchmarkRemoveEntry(b *testing.B) {
	c := New()
	// Pre-add jobs
	ids := make([]EntryID, 100)
	for i := 0; i < 100; i++ {
		id, _ := c.AddFunc("* * * * *", func(ctx context.Context) error {
			return nil
		})
		ids[i] = id
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a fresh cron for each iteration
		c := New()
		for j := 0; j < 100; j++ {
			_, _ = c.AddFunc("* * * * *", func(ctx context.Context) error {
				return nil
			})
		}
		// Remove middle entry
		c.removeEntry(EntryID(50))
	}
}
