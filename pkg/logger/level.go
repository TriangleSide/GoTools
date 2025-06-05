package logger

import (
	"fmt"
	"strings"
)

// LogLevel represents the various log levels.
type LogLevel int

const (
	LevelError LogLevel = iota
	LevelWarn
	LevelInfo
	LevelDebug
	LevelTrace
)

// appLogLevel is the configured log level for the application.
var appLogLevel = LevelInfo

// SetLevel sets the application log level.
func SetLevel(level LogLevel) {
	lock.Lock()
	defer lock.Unlock()
	appLogLevel = level
}

// GetLevel returns the application's log level.
func GetLevel() LogLevel {
	lock.RLock()
	defer lock.RUnlock()
	return appLogLevel
}

// String converts a LogLevel to its string representation.
func (l LogLevel) String() string {
	switch l {
	case LevelError:
		return "ERROR"
	case LevelWarn:
		return "WARN"
	case LevelInfo:
		return "INFO"
	case LevelDebug:
		return "DEBUG"
	case LevelTrace:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel parses a string into a LogLevel.
func ParseLevel(level string) (LogLevel, error) {
	switch strings.ToUpper(level) {
	case "ERROR":
		return LevelError, nil
	case "WARN":
		return LevelWarn, nil
	case "INFO":
		return LevelInfo, nil
	case "DEBUG":
		return LevelDebug, nil
	case "TRACE":
		return LevelTrace, nil
	default:
		return LevelError, fmt.Errorf("invalid log level: %s", level)
	}
}

// MarshalText serializes the LogLevel into text.
func (l LogLevel) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// UnmarshalText deserializes text into a LogLevel.
func (l *LogLevel) UnmarshalText(text []byte) error {
	levelStr := string(text)
	level, err := ParseLevel(levelStr)
	if err != nil {
		return err
	}
	*l = level
	return nil
}
