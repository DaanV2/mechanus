package logging

import (
	"context"

	"github.com/charmbracelet/log"
)

// With returns a logger with the given key-value pairs attached,
// based on the logger extracted from the provided context.
func With(ctx context.Context, keyvals ...interface{}) *log.Logger {
	return From(ctx).With(keyvals...)
}

// WithPrefix returns a logger with the given prefix attached,
// based on the logger extracted from the provided context.
func WithPrefix(ctx context.Context, prefix string) *log.Logger {
	return From(ctx).WithPrefix(prefix)
}

// Debug logs a message at the debug level using the logger from the context.
func Debug(ctx context.Context, msg interface{}, keyvals ...interface{}) {
	From(ctx).Debug(msg, keyvals...)
}

// Info logs a message at the info level using the logger from the context.
func Info(ctx context.Context, msg interface{}, keyvals ...interface{}) {
	From(ctx).Info(msg, keyvals...)
}

// Warn logs a message at the warning level using the logger from the context.
func Warn(ctx context.Context, msg interface{}, keyvals ...interface{}) {
	From(ctx).Warn(msg, keyvals...)
}

// Error logs a message at the error level using the logger from the context.
func Error(ctx context.Context, msg interface{}, keyvals ...interface{}) {
	From(ctx).Error(msg, keyvals...)
}

// Debugf logs a formatted message at the debug level using the logger from the context.
func Debugf(ctx context.Context, format string, args ...interface{}) {
	From(ctx).Debugf(format, args...)
}

// Infof logs a formatted message at the info level using the logger from the context.
func Infof(ctx context.Context, format string, args ...interface{}) {
	From(ctx).Infof(format, args...)
}

// Warnf logs a formatted message at the warning level using the logger from the context.
func Warnf(ctx context.Context, format string, args ...interface{}) {
	From(ctx).Warnf(format, args...)
}

// Errorf logs a formatted message at the error level using the logger from the context.
func Errorf(ctx context.Context, format string, args ...interface{}) {
	From(ctx).Errorf(format, args...)
}
