package logger

var (
	// appEntry is the entry for the entire application.
	appEntry Logger = &entry{}
)

// LogFn is used by the Logger and is invoked selectively when the log level is allowed.
type LogFn func() []any

// Logger describes log functions.
type Logger interface {
	Panic(args ...any)
	Panicf(format string, args ...any)
	PanicFn(fn LogFn)
	Error(args ...any)
	Errorf(format string, args ...any)
	ErrorFn(fn LogFn)
	Warn(args ...any)
	Warnf(format string, args ...any)
	WarnFn(fn LogFn)
	Info(args ...any)
	Infof(format string, args ...any)
	InfoFn(fn LogFn)
	Debug(args ...any)
	Debugf(format string, args ...any)
	DebugFn(fn LogFn)
	Trace(args ...any)
	Tracef(format string, args ...any)
	TraceFn(fn LogFn)
}

// Panic logs a message at panic level and then panics.
func Panic(args ...any) {
	appEntry.Panic(args...)
}

// Panicf logs a formatted message at panic level and then panics.
func Panicf(format string, args ...any) {
	appEntry.Panicf(format, args...)
}

// PanicFn logs a message at panic level using a deferred function and then panics.
func PanicFn(fn LogFn) {
	appEntry.PanicFn(fn)
}

// Error logs a message at error level.
func Error(args ...any) {
	appEntry.Error(args...)
}

// Errorf logs a formatted message at error level.
func Errorf(format string, args ...any) {
	appEntry.Errorf(format, args...)
}

// ErrorFn logs a message at error level using a deferred function.
func ErrorFn(fn LogFn) {
	appEntry.ErrorFn(fn)
}

// Warn logs a message at warning level.
func Warn(args ...any) {
	appEntry.Warn(args...)
}

// Warnf logs a formatted message at warning level.
func Warnf(format string, args ...any) {
	appEntry.Warnf(format, args...)
}

// WarnFn logs a message at warning level using a deferred function.
func WarnFn(fn LogFn) {
	appEntry.WarnFn(fn)
}

// Info logs a message at info level.
func Info(args ...any) {
	appEntry.Info(args...)
}

// Infof logs a formatted message at info level.
func Infof(format string, args ...any) {
	appEntry.Infof(format, args...)
}

// InfoFn logs a message at info level using a deferred function.
func InfoFn(fn LogFn) {
	appEntry.InfoFn(fn)
}

// Debug logs a message at debug level.
func Debug(args ...any) {
	appEntry.Debug(args...)
}

// Debugf logs a formatted message at debug level.
func Debugf(format string, args ...any) {
	appEntry.Debugf(format, args...)
}

// DebugFn logs a message at debug level using a deferred function.
func DebugFn(fn LogFn) {
	appEntry.DebugFn(fn)
}

// Trace logs a message at trace level.
func Trace(args ...any) {
	appEntry.Trace(args...)
}

// Tracef logs a formatted message at trace level.
func Tracef(format string, args ...any) {
	appEntry.Tracef(format, args...)
}

// TraceFn logs a message at trace level using a deferred function.
func TraceFn(fn LogFn) {
	appEntry.TraceFn(fn)
}
