package logger

import (
	"bytes"
	"log"
	"sync"

	"github.com/flc1125/go-cron/v4"
)

type Buffer struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func NewBuffer() *Buffer {
	return &Buffer{}
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *Buffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func (b *Buffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf.Reset()
}

func NewBufferLogger(buffer *Buffer) cron.Logger {
	return cron.VerbosePrintfLogger(log.New(buffer, "", log.LstdFlags))
}
