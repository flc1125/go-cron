package recovery

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/flc1125/go-cron/v4"
	"github.com/stretchr/testify/assert"
)

func newBufferLogger() (*bytes.Buffer, cron.Logger) {
	buf := new(bytes.Buffer)
	return buf, cron.VerbosePrintfLogger(log.New(buf, "", log.LstdFlags))
}

type panicJob struct{}

func (p panicJob) Run(context.Context) error {
	panic("YOLO")
}

func TestRecovery(t *testing.T) {
	buf, logger := newBufferLogger()
	recovery := New(
		WithLogger(logger),
	)

	assert.NotPanics(t, func() {
		_ = recovery(cron.JobFunc(func(context.Context) error {
			panic("YOLO")
		})).Run(t.Context())
	})

	assert.True(t, strings.Contains(buf.String(), "YOLO"))
}

func TestRecovery_FuncPanic(t *testing.T) {
	buf, logger := newBufferLogger()
	c := cron.New(
		cron.WithSeconds(),
		cron.WithMiddleware(
			New(
				WithLogger(logger),
			),
		),
	)
	c.Start()
	defer c.Stop()

	_, err := c.AddFunc("* * * * * ?", func(context.Context) error {
		panic("YOLO")
	})
	assert.NoError(t, err)

	time.Sleep(time.Second)
	assert.True(t, strings.Contains(buf.String(), "YOLO"))
}

func TestRecovery_JobPanic(t *testing.T) {
	buf, logger := newBufferLogger()
	c := cron.New(
		cron.WithSeconds(),
		cron.WithMiddleware(
			New(
				WithLogger(logger),
			),
		),
	)
	c.Start()
	defer c.Stop()

	_, err := c.AddJob("* * * * * ?", panicJob{})
	assert.NoError(t, err)

	time.Sleep(time.Second)
	assert.True(t, strings.Contains(buf.String(), "YOLO"))
}

func TestRecovery_ChainPanic(t *testing.T) {
	t.Run("default panic exits job", func(*testing.T) {
		assert.Panics(t, func() {
			_ = cron.Chain()(panicJob{}).Run(t.Context())
		})
	})

	t.Run("recovering job wrapper recovers", func(*testing.T) {
		buf, logger := newBufferLogger()
		assert.NotPanics(t, func() {
			_ = cron.Chain(
				New(WithLogger(logger)),
			)(panicJob{}).Run(t.Context())
		})
		assert.True(t, strings.Contains(buf.String(), "YOLO"))
	})
}
