package logger

import (
	"bytes"
	"log"

	"github.com/flc1125/go-cron/v4"
)

func NewBufferLogger(buf *bytes.Buffer) cron.Logger {
	return cron.PrintfLogger(log.New(buf, "", log.LstdFlags))
}
