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
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	FatalFn(fn LogFn)
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

func Panic(args ...any) {
	appEntry.Panic(args...)
}

func Panicf(format string, args ...any) {
	appEntry.Panicf(format, args...)
}

func PanicFn(fn LogFn) {
	appEntry.PanicFn(fn)
}

func Fatal(args ...any) {
	appEntry.Fatal(args...)
}

func Fatalf(format string, args ...any) {
	appEntry.Fatalf(format, args...)
}

func FatalFn(fn LogFn) {
	appEntry.FatalFn(fn)
}

func Error(args ...any) {
	appEntry.Error(args...)
}

func Errorf(format string, args ...any) {
	appEntry.Errorf(format, args...)
}

func ErrorFn(fn LogFn) {
	appEntry.ErrorFn(fn)
}

func Warn(args ...any) {
	appEntry.Warn(args...)
}

func Warnf(format string, args ...any) {
	appEntry.Warnf(format, args...)
}

func WarnFn(fn LogFn) {
	appEntry.WarnFn(fn)
}

func Info(args ...any) {
	appEntry.Info(args...)
}

func Infof(format string, args ...any) {
	appEntry.Infof(format, args...)
}

func InfoFn(fn LogFn) {
	appEntry.InfoFn(fn)
}

func Debug(args ...any) {
	appEntry.Debug(args...)
}

func Debugf(format string, args ...any) {
	appEntry.Debugf(format, args...)
}

func DebugFn(fn LogFn) {
	appEntry.DebugFn(fn)
}

func Trace(args ...any) {
	appEntry.Trace(args...)
}

func Tracef(format string, args ...any) {
	appEntry.Tracef(format, args...)
}

func TraceFn(fn LogFn) {
	appEntry.TraceFn(fn)
}
