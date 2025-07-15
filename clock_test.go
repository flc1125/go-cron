package cron

import (
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
