package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
)

var (
	appLogger = log.New(os.Stdout, "", 0)
)

func SetOutput(out io.Writer) {
	appLogger.SetOutput(out)
}

type LogFn func() []any

func Panic(ctx context.Context, args ...any) {
	appLogger.Panicln(formatLog(ctx, fmt.Sprint(args...)))
}

func Panicf(ctx context.Context, format string, args ...any) {
	appLogger.Panicln(formatLog(ctx, fmt.Sprintf(format, args...)))
}

func PanicFn(ctx context.Context, fn LogFn) {
	appLogger.Panicln(formatLog(ctx, fmt.Sprint(fn()...)))
}

func Fatal(ctx context.Context, args ...any) {
	appLogger.Fatalln(formatLog(ctx, fmt.Sprint(args...)))
}

func Fatalf(ctx context.Context, format string, args ...any) {
	appLogger.Fatalln(formatLog(ctx, fmt.Sprintf(format, args...)))
}

func FatalFn(ctx context.Context, fn LogFn) {
	appLogger.Fatalln(formatLog(ctx, fmt.Sprint(fn()...)))
}

func Error(ctx context.Context, args ...any) {
	if appLogLevel >= LevelError {
		appLogger.Println(formatLog(ctx, fmt.Sprint(args...)))
	}
}

func Errorf(ctx context.Context, format string, args ...any) {
	if appLogLevel >= LevelError {
		appLogger.Println(formatLog(ctx, fmt.Sprintf(format, args...)))
	}
}

func ErrorFn(ctx context.Context, fn LogFn) {
	if appLogLevel >= LevelError {
		appLogger.Println(formatLog(ctx, fmt.Sprint(fn()...)))
	}
}

func Warn(ctx context.Context, args ...any) {
	if appLogLevel >= LevelWarn {
		appLogger.Println(formatLog(ctx, fmt.Sprint(args...)))
	}
}

func Warnf(ctx context.Context, format string, args ...any) {
	if appLogLevel >= LevelWarn {
		appLogger.Println(formatLog(ctx, fmt.Sprintf(format, args...)))
	}
}

func WarnFn(ctx context.Context, fn LogFn) {
	if appLogLevel >= LevelWarn {
		appLogger.Println(formatLog(ctx, fmt.Sprint(fn()...)))
	}
}

func Info(ctx context.Context, args ...any) {
	if appLogLevel >= LevelInfo {
		appLogger.Println(formatLog(ctx, fmt.Sprint(args...)))
	}
}

func Infof(ctx context.Context, format string, args ...any) {
	if appLogLevel >= LevelInfo {
		appLogger.Println(formatLog(ctx, fmt.Sprintf(format, args...)))
	}
}

func InfoFn(ctx context.Context, fn LogFn) {
	if appLogLevel >= LevelInfo {
		appLogger.Println(formatLog(ctx, fmt.Sprint(fn()...)))
	}
}

func Debug(ctx context.Context, args ...any) {
	if appLogLevel >= LevelDebug {
		appLogger.Println(formatLog(ctx, fmt.Sprint(args...)))
	}
}

func Debugf(ctx context.Context, format string, args ...any) {
	if appLogLevel >= LevelDebug {
		appLogger.Println(formatLog(ctx, fmt.Sprintf(format, args...)))
	}
}

func DebugFn(ctx context.Context, fn LogFn) {
	if appLogLevel >= LevelDebug {
		appLogger.Println(formatLog(ctx, fmt.Sprint(fn()...)))
	}
}

func Trace(ctx context.Context, args ...any) {
	if appLogLevel >= LevelTrace {
		appLogger.Println(formatLog(ctx, fmt.Sprint(args...)))
	}
}

func Tracef(ctx context.Context, format string, args ...any) {
	if appLogLevel >= LevelTrace {
		appLogger.Println(formatLog(ctx, fmt.Sprintf(format, args...)))
	}
}

func TraceFn(ctx context.Context, fn LogFn) {
	if appLogLevel >= LevelTrace {
		appLogger.Println(formatLog(ctx, fmt.Sprint(fn()...)))
	}
}
