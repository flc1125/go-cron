package recovery

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flc1125/go-cron/v4"
	"github.com/flc1125/go-cron/v4/internal/logger"
)

type panicJob struct{}

func (p panicJob) Run(context.Context) error {
	panic("YOLO")
}

func TestRecovery(t *testing.T) {
	var buf bytes.Buffer
	recovery := New(
		WithLogger(logger.NewBufferLogger(&buf)),
	)

	assert.NotPanics(t, func() {
		assert.NoError(t, recovery(cron.JobFunc(func(context.Context) error {
			panic("YOLO")
		})).Run(context.Background()))
	})

	assert.True(t, strings.Contains(buf.String(), "YOLO"))
}

func TestRecovery_FuncPanic(t *testing.T) {
	var buf bytes.Buffer
	c := cron.New(
		cron.WithSeconds(),
		cron.WithMiddleware(
			New(
				WithLogger(logger.NewBufferLogger(&buf)),
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
	var buf bytes.Buffer
	c := cron.New(
		cron.WithSeconds(),
		cron.WithMiddleware(
			New(
				WithLogger(logger.NewBufferLogger(&buf)),
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
			assert.NotNil(t,
				cron.Chain()(panicJob{}).Run(context.Background()),
			)
		})
	})

	t.Run("recovering job wrapper recovers", func(*testing.T) {
		var buf bytes.Buffer

		assert.NotPanics(t, func() {
			assert.NoError(t,
				cron.Chain(
					New(WithLogger(logger.NewBufferLogger(&buf))),
				)(panicJob{}).Run(context.Background()),
			)
		})
		assert.True(t, strings.Contains(buf.String(), "YOLO"))
	})
}
