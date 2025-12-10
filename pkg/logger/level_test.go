package logger_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/logger"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestLogLevelString_AllLevels_ReturnsCorrectStringRepresentation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		level    logger.LogLevel
		expected string
	}{
		{"Error", logger.LevelError, "ERROR"},
		{"Warn", logger.LevelWarn, "WARN"},
		{"Info", logger.LevelInfo, "INFO"},
		{"Debug", logger.LevelDebug, "DEBUG"},
		{"Trace", logger.LevelTrace, "TRACE"},
		{"Unknown", logger.LogLevel(999), "UNKNOWN"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			actual := testCase.level.String()
			assert.Equals(t, actual, testCase.expected)
		})
	}
}

func TestParseLevel_ValidStrings_ReturnsCorrectLogLevel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected logger.LogLevel
	}{
		{"Error", "ERROR", logger.LevelError},
		{"Warn", "WARN", logger.LevelWarn},
		{"Info", "INFO", logger.LevelInfo},
		{"Debug", "DEBUG", logger.LevelDebug},
		{"Trace", "TRACE", logger.LevelTrace},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			level, err := logger.ParseLevel(testCase.input)
			assert.NoError(t, err)
			assert.Equals(t, level, testCase.expected)
		})
	}
}

func TestParseLevel_InvalidString_ReturnsError(t *testing.T) {
	t.Parallel()

	level, err := logger.ParseLevel("INVALID")
	assert.Equals(t, level, logger.LevelError)
	assert.Error(t, err)
}

func TestLogLevelMarshalText_AllLevels_ReturnsCorrectTextRepresentation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		level    logger.LogLevel
		expected string
	}{
		{"Error", logger.LevelError, "ERROR"},
		{"Warn", logger.LevelWarn, "WARN"},
		{"Info", logger.LevelInfo, "INFO"},
		{"Debug", logger.LevelDebug, "DEBUG"},
		{"Trace", logger.LevelTrace, "TRACE"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			marshaled, err := testCase.level.MarshalText()
			assert.NoError(t, err)
			assert.Equals(t, string(marshaled), testCase.expected)
		})
	}
}

func TestLogLevelUnmarshalText_ValidStrings_ReturnsCorrectLogLevel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected logger.LogLevel
	}{
		{"Error", "ERROR", logger.LevelError},
		{"Warn", "WARN", logger.LevelWarn},
		{"Info", "INFO", logger.LevelInfo},
		{"Debug", "DEBUG", logger.LevelDebug},
		{"Trace", "TRACE", logger.LevelTrace},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var level logger.LogLevel
			err := level.UnmarshalText([]byte(testCase.input))
			assert.NoError(t, err)
			assert.Equals(t, level, testCase.expected)
		})
	}
}

func TestLogLevelUnmarshalText_InvalidString_ReturnsError(t *testing.T) {
	t.Parallel()

	var level logger.LogLevel
	err := level.UnmarshalText([]byte("INVALID"))
	assert.Equals(t, level, logger.LevelError)
	assert.Error(t, err)
}

func TestParseLevel_LowercaseInput_ReturnsCorrectLogLevel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected logger.LogLevel
	}{
		{"LowercaseError", "error", logger.LevelError},
		{"LowercaseWarn", "warn", logger.LevelWarn},
		{"LowercaseInfo", "info", logger.LevelInfo},
		{"LowercaseDebug", "debug", logger.LevelDebug},
		{"LowercaseTrace", "trace", logger.LevelTrace},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			level, err := logger.ParseLevel(testCase.input)
			assert.NoError(t, err)
			assert.Equals(t, level, testCase.expected)
		})
	}
}

func TestParseLevel_MixedCaseInput_ReturnsCorrectLogLevel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected logger.LogLevel
	}{
		{"MixedCaseError", "Error", logger.LevelError},
		{"MixedCaseWarn", "WaRn", logger.LevelWarn},
		{"MixedCaseInfo", "iNfO", logger.LevelInfo},
		{"MixedCaseDebug", "DeBuG", logger.LevelDebug},
		{"MixedCaseTrace", "tRaCe", logger.LevelTrace},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			level, err := logger.ParseLevel(testCase.input)
			assert.NoError(t, err)
			assert.Equals(t, level, testCase.expected)
		})
	}
}

func TestParseLevel_EmptyString_ReturnsError(t *testing.T) {
	t.Parallel()

	level, err := logger.ParseLevel("")
	assert.Equals(t, level, logger.LevelError)
	assert.Error(t, err)
}

func TestLogLevelMarshalText_UnknownLevel_ReturnsUnknown(t *testing.T) {
	t.Parallel()

	level := logger.LogLevel(999)
	marshaled, err := level.MarshalText()
	assert.NoError(t, err)
	assert.Equals(t, string(marshaled), "UNKNOWN")
}

func TestLogLevelUnmarshalText_EmptyString_ReturnsError(t *testing.T) {
	t.Parallel()

	var level logger.LogLevel
	err := level.UnmarshalText([]byte(""))
	assert.Equals(t, level, logger.LevelError)
	assert.Error(t, err)
}

func TestLogLevelUnmarshalText_LowercaseInput_ReturnsCorrectLogLevel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected logger.LogLevel
	}{
		{"LowercaseError", "error", logger.LevelError},
		{"LowercaseWarn", "warn", logger.LevelWarn},
		{"LowercaseInfo", "info", logger.LevelInfo},
		{"LowercaseDebug", "debug", logger.LevelDebug},
		{"LowercaseTrace", "trace", logger.LevelTrace},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var level logger.LogLevel
			err := level.UnmarshalText([]byte(testCase.input))
			assert.NoError(t, err)
			assert.Equals(t, level, testCase.expected)
		})
	}
}

func TestSetLevel_ValidLevel_UpdatesLevel(t *testing.T) {
	t.Cleanup(func() {
		logger.SetLevel(logger.LevelInfo)
	})

	logger.SetLevel(logger.LevelDebug)
	assert.Equals(t, logger.GetLevel(), logger.LevelDebug)
}

func TestGetLevel_DefaultLevel_ReturnsInfo(t *testing.T) {
	t.Cleanup(func() {
		logger.SetLevel(logger.LevelInfo)
	})

	logger.SetLevel(logger.LevelInfo)
	level := logger.GetLevel()
	assert.Equals(t, level, logger.LevelInfo)
}

func TestSetLevel_AllLevels_UpdatesLevelCorrectly(t *testing.T) {
	t.Cleanup(func() {
		logger.SetLevel(logger.LevelInfo)
	})

	testCases := []struct {
		name     string
		level    logger.LogLevel
		expected logger.LogLevel
	}{
		{"Error", logger.LevelError, logger.LevelError},
		{"Warn", logger.LevelWarn, logger.LevelWarn},
		{"Info", logger.LevelInfo, logger.LevelInfo},
		{"Debug", logger.LevelDebug, logger.LevelDebug},
		{"Trace", logger.LevelTrace, logger.LevelTrace},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			logger.SetLevel(testCase.level)
			assert.Equals(t, logger.GetLevel(), testCase.expected)
		})
	}
}

func TestLogLevelString_NegativeLevel_ReturnsUnknown(t *testing.T) {
	t.Parallel()

	level := logger.LogLevel(-1)
	assert.Equals(t, level.String(), "UNKNOWN")
}
