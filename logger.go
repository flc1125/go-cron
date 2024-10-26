package cron

import (
	"io/ioutil" //nolint:staticcheck // todo: Waiting for refactoring
	"log"
	"os"
	"strings"
	"time"
)

// DefaultLogger is used by Cron if none is specified.
var DefaultLogger = PrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))

// DiscardLogger can be used by callers to discard all log messages.
var DiscardLogger = PrintfLogger(log.New(ioutil.Discard, "", 0))

// Logger is the interface used in this package for logging, so that any backend
// can be plugged in. It is a subset of the github.com/go-logr/logr interface.
type Logger interface {
	// Info logs routine messages about cron's operation.
	Info(msg string, keysAndValues ...any)
	// Error logs an error condition.
	Error(err error, msg string, keysAndValues ...any)
}

// PrintfLogger wraps a Printf-based logger (such as the standard library "log")
// into an implementation of the Logger interface which logs errors only.
func PrintfLogger(l interface{ Printf(string, ...any) }) Logger {
	return printfLogger{l, false}
}

// VerbosePrintfLogger wraps a Printf-based logger (such as the standard library
// "log") into an implementation of the Logger interface which logs everything.
func VerbosePrintfLogger(l interface{ Printf(string, ...any) }) Logger {
	return printfLogger{l, true}
}

type printfLogger struct {
	logger  interface{ Printf(string, ...any) }
	logInfo bool
}

func (pl printfLogger) Info(msg string, keysAndValues ...any) {
	if pl.logInfo {
		keysAndValues = formatTimes(keysAndValues)
		pl.logger.Printf(
			formatString(len(keysAndValues)),
			append([]any{msg}, keysAndValues...)...)
	}
}

func (pl printfLogger) Error(err error, msg string, keysAndValues ...any) {
	keysAndValues = formatTimes(keysAndValues)
	pl.logger.Printf(
		formatString(len(keysAndValues)+2),
		append([]any{msg, "error", err}, keysAndValues...)...)
}

// formatString returns a logfmt-like format string for the number of
// key/values.
func formatString(numKeysAndValues int) string {
	var sb strings.Builder
	sb.WriteString("%s")
	if numKeysAndValues > 0 {
		sb.WriteString(", ")
	}
	for i := 0; i < numKeysAndValues/2; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("%v=%v")
	}
	return sb.String()
}

// formatTimes formats any time.Time values as RFC3339.
func formatTimes(keysAndValues []any) []any {
	var formattedArgs []any // nolint:prealloc
	for _, arg := range keysAndValues {
		if t, ok := arg.(time.Time); ok {
			arg = t.Format(time.RFC3339)
		}
		formattedArgs = append(formattedArgs, arg)
	}
	return formattedArgs
}
