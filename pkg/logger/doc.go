/*
Package logger provides context-based structured logging built on Go's slog package.

This package is useful when you need to propagate a logger through your application
via context and maintain consistent logging attributes across request boundaries. It
outputs JSON-formatted logs to stdout and supports configuring the log level through
the LOG_LEVEL environment variable.
*/
package logger
