package logger

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// SetOutput sets the writer for the application logger.
func SetOutput(out io.Writer) {
	lock.Lock()
	defer lock.Unlock()
	appLogger.SetOutput(out)
}

// FormatterFunc defines a function type that takes in a map of fields and a message string
// and returns a formatted log string. This allows for customizable log formatting.
type FormatterFunc func(fields map[string]any, msg string) string

// appLogFormatter holds the current log formatter function.
var appLogFormatter FormatterFunc = DefaultFormatter

// DefaultFormatter formats the log entry.
func DefaultFormatter(fields map[string]any, msg string) string {
	timestamp := time.Now().UTC().Format(time.DateTime)
	fieldsSb := strings.Builder{}
	for k, v := range fields {
		fieldsSb.WriteString(fmt.Sprintf("%s=%v ", k, v))
	}
	return fmt.Sprintf("%s %s%s", timestamp, fieldsSb.String(), msg)
}

// SetFormatter sets a custom log formatter function.
// It replaces the default log formatter with the provided one.
func SetFormatter(formatter FormatterFunc) {
	lock.Lock()
	defer lock.Unlock()
	appLogFormatter = formatter
}

// formatLog formats the log message using the fields in the context and the provided message.
func formatLog(fields map[string]any, msg string) string {
	lock.RLock()
	defer lock.RUnlock()
	return appLogFormatter(fields, msg)
}
