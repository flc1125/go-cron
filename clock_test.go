package cron

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestSystemClock(t *testing.T) {
	utc := time.UTC
	clock := SystemClock{location: utc}
	now := clock.Now()
	if now.Location() != utc {
		t.Errorf("expected timezone %v, got %v", utc, now.Location())
	}
	
	if time.Since(now) > time.Second {
		t.Error("SystemClock.Now() should return current time")
	}
}

func TestOffsetClock(t *testing.T) {
	baseClock := SystemClock{location: time.UTC}
	offset := 30 * time.Second
	clock := OffsetClock{
		base:   baseClock,
		offset: offset,
	}
	
	baseTime := baseClock.Now()
	offsetTime := clock.Now()
	
	expectedTime := baseTime.Add(offset)
	if offsetTime.Sub(expectedTime).Abs() > time.Millisecond {
		t.Errorf("expected offset time %v, got %v", expectedTime, offsetTime)
	}
}

func TestOffsetClockNegative(t *testing.T) {
	baseClock := SystemClock{location: time.UTC}
	offset := -1 * time.Minute
	clock := OffsetClock{
		base:   baseClock,
		offset: offset,
	}
	
	baseTime := baseClock.Now()
	offsetTime := clock.Now()
	
	expectedTime := baseTime.Add(offset)
	if offsetTime.Sub(expectedTime).Abs() > time.Millisecond {
		t.Errorf("expected offset time %v, got %v", expectedTime, offsetTime)
	}
}

func TestClockIntegration(t *testing.T) {
	offset := 30 * time.Second
	c := New(WithOffset(offset))
	
	offsetClock, ok := c.clock.(OffsetClock)
	if !ok {
		t.Fatalf("expected OffsetClock, got %T", c.clock)
	}
	
	if offsetClock.offset != offset {
		t.Errorf("expected offset %v, got %v", offset, offsetClock.offset)
	}
	
	now := c.now()
	systemNow := time.Now()
	
	expectedTime := systemNow.Add(offset)
	if now.Sub(expectedTime).Abs() > time.Second {
		t.Errorf("expected time around %v, got %v", expectedTime, now)
	}
}

func TestWithClock(t *testing.T) {
	clock := SystemClock{location: time.UTC}
	c := New(WithClock(clock))
	if c.clock != clock {
		t.Errorf("expected clock %v, got %v", clock, c.clock)
	}
}

func TestOffsetSpecIntegration(t *testing.T) {
	c := New(WithSeconds())
	
	_, err := c.AddFunc("OFFSET=30s 0 * * * * *", func(ctx context.Context) error {
		return nil
	})
	
	if err != nil {
		t.Errorf("expected success but got error: %v", err)
	}
	
	offsetClock, ok := c.clock.(OffsetClock)
	if !ok {
		t.Fatalf("expected OffsetClock after OFFSET= spec, got %T", c.clock)
	}
	
	if offsetClock.offset != 30*time.Second {
		t.Errorf("expected offset 30s, got %v", offsetClock.offset)
	}
}

func TestOffsetSpecErrors(t *testing.T) {
	c := New(WithSeconds())
	
	tests := []struct {
		spec string
		err  string
	}{
		{"OFFSET=invalid 0 * * * * *", "provided bad offset"},
		{"OFFSET= 0 * * * * *", "provided bad offset"},
		{"OFFSET=1x 0 * * * * *", "provided bad offset"},
	}
	
	for _, test := range tests {
		_, err := c.AddFunc(test.spec, func(ctx context.Context) error {
			return nil
		})
		if err == nil || !strings.Contains(err.Error(), test.err) {
			t.Errorf("%s => expected error containing %q, got %v", test.spec, test.err, err)
		}
	}
}
