package cron

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJob_NoopJob(t *testing.T) {
	assert.NoError(t, NoopJob{}.Run(t.Context()))
}
