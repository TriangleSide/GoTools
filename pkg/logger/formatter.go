package logger

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// logFormatterType defines a function type that takes in a map of fields and a message string
// and returns a formatted log string. This allows for customizable log formatting.
type logFormatterType func(fields map[string]any, msg string) string

// appLogFormatter holds the current log formatter function.
var appLogFormatter logFormatterType = defaultLogFormatter

// defaultLogFormatter formats the log entry.
func defaultLogFormatter(fields map[string]any, msg string) string {
	timestamp := time.Now().UTC().Format(time.DateTime)
	fieldsSb := strings.Builder{}
	for k, v := range fields {
		fieldsSb.WriteString(fmt.Sprintf("%s=%v ", k, v))
	}
	return fmt.Sprintf("%s %s%s", timestamp, fieldsSb.String(), msg)
}

// SetFormatter sets a custom log formatter function.
// It replaces the default log formatter with the provided one.
func SetFormatter(f logFormatterType) {
	appLogFormatter = f
}

// formatLog formats the log message using the fields in the context and the provided message.
func formatLog(ctx context.Context, msg string) string {
	if ctx == nil {
		return appLogFormatter(nil, msg)
	}
	fieldsNotCast := ctx.Value(contextKey)
	if fieldsNotCast == nil {
		return appLogFormatter(nil, msg)
	}
	fields, fieldsCastOk := fieldsNotCast.(map[string]any)
	if !fieldsCastOk {
		panic("The logger context fields is not the correct type.")
	}
	return appLogFormatter(fields, msg)
}
