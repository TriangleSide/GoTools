package logger

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"

	"intelligence/pkg/config"
)

var (
	logger *logrus.Logger
)

// init configures the applications logger.
func init() {
	logger = logrus.New()
	logger.SetOutput(os.Stdout)
	if conf, confErr := config.ProcessAndValidate[config.Logger](); confErr == nil {
		if level, parseLevelErr := logrus.ParseLevel(conf.LogLevel); parseLevelErr == nil {
			logger.SetLevel(level)
		} else {
			logger.WithError(parseLevelErr).Fatal("Failed to parse logger level.")
		}
	} else {
		logger.WithError(confErr).Fatal("Failed to parse logger config.")
	}
	logger.SetFormatter(&customFormatter{
		JSONFormatter: logrus.JSONFormatter{
			// UTC is enforced in the customFormatter.
			TimestampFormat: "2001-02-03 19:34:56",
			PrettyPrint:     logger.Level >= logrus.DebugLevel,
		},
	})
}

// logEntry is the internal method that returns the log entry pointer stored in the context.
func logEntry(ctx *context.Context) *logrus.Entry {
	logEntryValue := (*ctx).Value(logEntryContextKey)
	if logEntryValue != nil {
		if entry, ok := logEntryValue.(*logrus.Entry); ok {
			return entry
		} else {
			logger.Fatal("Log entry in the context does not match the expected type.")
		}
	}
	return logrus.NewEntry(logger)
}

// LogEntry returns a copy of the log entry stored in the context.
// To modify the fields in the log entry and pass those fields along with the context, use WithField.
func LogEntry(ctx context.Context) *logrus.Entry {
	return logEntry(&ctx).Dup()
}

// WithField adds a field to the log entry. The log entry is stored in the context and passed along with it.
func WithField(ctx *context.Context, key LogEntryKey, value any) *logrus.Entry {
	entry := logEntry(ctx)
	entry = entry.WithField(string(key), value)
	*ctx = context.WithValue(*ctx, logEntryContextKey, entry)
	return entry
}
