package cron

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithContext(t *testing.T) {
	type key struct{}
	ctx := context.WithValue(t.Context(), key{}, "value")
	c := New(WithContext(ctx))
	assert.Equal(t, "value", c.ctx.Value(key{}))
}

func TestWithLocation(t *testing.T) {
	c := New(WithLocation(time.UTC))
	assert.Equal(t, time.UTC, c.location)
}

func TestWithParser(t *testing.T) {
	parser := NewParser(Dow)
	c := New(WithParser(parser))
	assert.Equal(t, parser, c.parser)
}

func TestWithVerboseLogger(t *testing.T) {
	var buf syncWriter
	logger := log.New(&buf, "", log.LstdFlags)
	c := New(WithLogger(VerbosePrintfLogger(logger)))
	assert.Same(t, logger, c.logger.(printfLogger).logger)

	c.AddFunc("@every 1s", func(context.Context) error { return nil }) //nolint:errcheck
	c.Start()
	time.Sleep(OneSecond)
	c.Stop()
	out := buf.String()
	assert.Contains(t, out, "schedule,")
	assert.Contains(t, out, "run,")
}
