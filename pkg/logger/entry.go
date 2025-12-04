package logger

import (
	"fmt"
	"log"
	"os"
)

var (
	// appLogger is the logger instance for the entire application.
	appLogger = log.New(os.Stdout, "", 0)
)

// entry implements the Logger interface.
// It logs with the available fields.
type entry struct {
	fields map[string]any
}

func (l *entry) Panic(args ...any) {
	lock.RLock()
	defer lock.RUnlock()
	appLogger.Panicln(formatLog(l.fields, fmt.Sprint(args...)))
}

func (l *entry) Panicf(format string, args ...any) {
	lock.RLock()
	defer lock.RUnlock()
	appLogger.Panicln(formatLog(l.fields, fmt.Sprintf(format, args...)))
}

func (l *entry) PanicFn(fn LogFn) {
	lock.RLock()
	defer lock.RUnlock()
	appLogger.Panicln(formatLog(l.fields, fmt.Sprint(fn()...)))
}

func (l *entry) Error(args ...any) {
	lock.RLock()
	defer lock.RUnlock()
	if appLogLevel >= LevelError {
		appLogger.Println(formatLog(l.fields, fmt.Sprint(args...)))
	}
}

func (l *entry) Errorf(format string, args ...any) {
	lock.RLock()
	defer lock.RUnlock()
	if appLogLevel >= LevelError {
		appLogger.Println(formatLog(l.fields, fmt.Sprintf(format, args...)))
	}
}

func (l *entry) ErrorFn(fn LogFn) {
	lock.RLock()
	defer lock.RUnlock()
	if appLogLevel >= LevelError {
		appLogger.Println(formatLog(l.fields, fmt.Sprint(fn()...)))
	}
}

func (l *entry) Warn(args ...any) {
	lock.RLock()
	defer lock.RUnlock()
	if appLogLevel >= LevelWarn {
		appLogger.Println(formatLog(l.fields, fmt.Sprint(args...)))
	}
}

func (l *entry) Warnf(format string, args ...any) {
	lock.RLock()
	defer lock.RUnlock()
	if appLogLevel >= LevelWarn {
		appLogger.Println(formatLog(l.fields, fmt.Sprintf(format, args...)))
	}
}

func (l *entry) WarnFn(fn LogFn) {
	lock.RLock()
	defer lock.RUnlock()
	if appLogLevel >= LevelWarn {
		appLogger.Println(formatLog(l.fields, fmt.Sprint(fn()...)))
	}
}

func (l *entry) Info(args ...any) {
	lock.RLock()
	defer lock.RUnlock()
	if appLogLevel >= LevelInfo {
		appLogger.Println(formatLog(l.fields, fmt.Sprint(args...)))
	}
}

func (l *entry) Infof(format string, args ...any) {
	lock.RLock()
	defer lock.RUnlock()
	if appLogLevel >= LevelInfo {
		appLogger.Println(formatLog(l.fields, fmt.Sprintf(format, args...)))
	}
}

func (l *entry) InfoFn(fn LogFn) {
	lock.RLock()
	defer lock.RUnlock()
	if appLogLevel >= LevelInfo {
		appLogger.Println(formatLog(l.fields, fmt.Sprint(fn()...)))
	}
}

func (l *entry) Debug(args ...any) {
	lock.RLock()
	defer lock.RUnlock()
	if appLogLevel >= LevelDebug {
		appLogger.Println(formatLog(l.fields, fmt.Sprint(args...)))
	}
}

func (l *entry) Debugf(format string, args ...any) {
	lock.RLock()
	defer lock.RUnlock()
	if appLogLevel >= LevelDebug {
		appLogger.Println(formatLog(l.fields, fmt.Sprintf(format, args...)))
	}
}

func (l *entry) DebugFn(fn LogFn) {
	lock.RLock()
	defer lock.RUnlock()
	if appLogLevel >= LevelDebug {
		appLogger.Println(formatLog(l.fields, fmt.Sprint(fn()...)))
	}
}

func (l *entry) Trace(args ...any) {
	lock.RLock()
	defer lock.RUnlock()
	if appLogLevel >= LevelTrace {
		appLogger.Println(formatLog(l.fields, fmt.Sprint(args...)))
	}
}

func (l *entry) Tracef(format string, args ...any) {
	lock.RLock()
	defer lock.RUnlock()
	if appLogLevel >= LevelTrace {
		appLogger.Println(formatLog(l.fields, fmt.Sprintf(format, args...)))
	}
}

func (l *entry) TraceFn(fn LogFn) {
	lock.RLock()
	defer lock.RUnlock()
	if appLogLevel >= LevelTrace {
		appLogger.Println(formatLog(l.fields, fmt.Sprint(fn()...)))
	}
}
