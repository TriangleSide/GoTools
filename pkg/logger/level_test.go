package logger_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/logger"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestLogLevel(t *testing.T) {
	t.Run("when converting a log level to string representation", func(t *testing.T) {
		testCases := []struct {
			level    logger.LogLevel
			expected string
		}{
			{logger.LevelError, "ERROR"},
			{logger.LevelWarn, "WARN"},
			{logger.LevelInfo, "INFO"},
			{logger.LevelDebug, "DEBUG"},
			{logger.LevelTrace, "TRACE"},
			{logger.LogLevel(999), "UNKNOWN"},
		}

		for _, testCase := range testCases {
			actual := testCase.level.String()
			assert.Equals(t, actual, testCase.expected)
		}
	})

	t.Run("when parsing a string to a log level", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected logger.LogLevel
			hasError bool
		}{
			{"ERROR", logger.LevelError, false},
			{"WARN", logger.LevelWarn, false},
			{"INFO", logger.LevelInfo, false},
			{"DEBUG", logger.LevelDebug, false},
			{"TRACE", logger.LevelTrace, false},
			{"INVALID", logger.LevelError, true},
		}

		for _, testCase := range testCases {
			level, err := logger.ParseLevel(testCase.input)
			assert.Equals(t, level, testCase.expected)
			if testCase.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		}
	})

	t.Run("when marshaling text representation of log levels", func(t *testing.T) {
		testCases := []struct {
			level    logger.LogLevel
			expected string
		}{
			{logger.LevelError, "ERROR"},
			{logger.LevelWarn, "WARN"},
			{logger.LevelInfo, "INFO"},
			{logger.LevelDebug, "DEBUG"},
			{logger.LevelTrace, "TRACE"},
		}

		for _, testCase := range testCases {
			marshaled, err := testCase.level.MarshalText()
			assert.NoError(t, err)
			assert.Equals(t, string(marshaled), testCase.expected)
		}
	})

	t.Run("when unmarshalling text to a log level", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected logger.LogLevel
			hasError bool
		}{
			{"ERROR", logger.LevelError, false},
			{"WARN", logger.LevelWarn, false},
			{"INFO", logger.LevelInfo, false},
			{"DEBUG", logger.LevelDebug, false},
			{"TRACE", logger.LevelTrace, false},
			{"INVALID", logger.LevelError, true},
		}

		for _, testCase := range testCases {
			var level logger.LogLevel
			err := level.UnmarshalText([]byte(testCase.input))
			assert.Equals(t, level, testCase.expected)
			if testCase.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		}
	})
}
