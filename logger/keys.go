package logger

// logEntryContextKeyType is the key used in the context for the log entry.
type logEntryContextKeyType string

const (
	// This is used to ensure there's no collision for the context key.
	logEntryContextKey logEntryContextKeyType = "logEntry"
)

// LogEntryKey is the key used for log entry field values.
// The declared type ensures the consistent naming of keys across the program.
type LogEntryKey string
