package recovery

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/flc1125/go-cron/crontest/v4/logger"
	"github.com/flc1125/go-cron/v4"
	"github.com/stretchr/testify/assert"
)

type panicJob struct{}

func (p panicJob) Run(context.Context) error {
	panic("YOLO")
}

func TestRecovery(t *testing.T) {
	buf := logger.NewBuffer()
	recovery := New(
		WithLogger(logger.NewBufferLogger(buf)),
	)

	assert.NotPanics(t, func() {
		_ = recovery(cron.JobFunc(func(context.Context) error {
			panic("YOLO")
		})).Run(context.Background())
	})

	assert.True(t, strings.Contains(buf.String(), "YOLO"))
}

func TestRecovery_FuncPanic(t *testing.T) {
	buf := logger.NewBuffer()
	c := cron.New(
		cron.WithSeconds(),
		cron.WithMiddleware(
			New(
				WithLogger(logger.NewBufferLogger(buf)),
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
	buf := logger.NewBuffer()
	c := cron.New(
		cron.WithSeconds(),
		cron.WithMiddleware(
			New(
				WithLogger(logger.NewBufferLogger(buf)),
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
			_ = cron.Chain()(panicJob{}).Run(context.Background())
		})
	})

	t.Run("recovering job wrapper recovers", func(*testing.T) {
		var buf logger.Buffer
		assert.NotPanics(t, func() {
			_ = cron.Chain(
				New(WithLogger(logger.NewBufferLogger(&buf))),
			)(panicJob{}).Run(context.Background())
		})
		assert.True(t, strings.Contains(buf.String(), "YOLO"))
	})
}
