package logger

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/TriangleSide/GoBase/pkg/config"
	"github.com/TriangleSide/GoBase/pkg/config/envprocessor"
)

const (
	defaultLogLevel = logrus.InfoLevel
)

var (
	// logger is the applications logger.
	logger *logrus.Logger
)

// init configures the applications logger with default settings.
func init() {
	logger = logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(defaultLogLevel)
}

// Config is configured by the Option functions.
type Config struct {
	configProvider func() (*config.Logger, error)
}

// Option sets values on the Config.
type Option func(*Config) error

// WithConfigProvider overwrites the default config provider.
func WithConfigProvider(provider func() (*config.Logger, error)) Option {
	return func(c *Config) error {
		c.configProvider = provider
		return nil
	}
}

// MustConfigure parses the logger conf and configures the application logger.
func MustConfigure(opts ...Option) {
	cfg := Config{
		configProvider: func() (*config.Logger, error) {
			return envprocessor.ProcessAndValidate[config.Logger]()
		},
	}

	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			panic(fmt.Sprintf("Failed to set the options for the logger (%s).", err.Error()))
		}
	}

	logger.SetFormatter(&UTCFormatter{
		Next: &logrus.JSONFormatter{
			TimestampFormat: "2001-02-03 19:34:56",
			PrettyPrint:     false,
		},
	})

	envConf, err := cfg.configProvider()
	if err != nil {
		panic(fmt.Sprintf("Failed to get logger config (%s).", err.Error()))
	}

	level, err := logrus.ParseLevel(envConf.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse the log level (%s).", err.Error()))
	}
	logger.SetLevel(level)
}

// logEntry is an internal method that returns a log entry pointer stored in the context.
func logEntry(ctx *context.Context) *logrus.Entry {
	logEntryValue := (*ctx).Value(logEntryContextKey)
	if logEntryValue != nil {
		return logEntryValue.(*logrus.Entry)
	}
	return logrus.NewEntry(logger)
}

// LogEntry returns a copy of the log entry stored in the context.
// To modify the fields in the log entry and pass those fields along with the context, use WithField.
func LogEntry(ctx context.Context) *logrus.Entry {
	return logEntry(&ctx).Dup()
}

// WithField adds a field to the log entry stored in the context.
func WithField(ctx *context.Context, key LogEntryKey, value any) *logrus.Entry {
	entry := logEntry(ctx)
	entry = entry.WithField(string(key), value)
	*ctx = context.WithValue(*ctx, logEntryContextKey, entry)
	return entry
}
