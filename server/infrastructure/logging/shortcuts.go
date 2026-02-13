package logging

import (
	"context"

	"github.com/charmbracelet/log"
	"go.opentelemetry.io/otel/trace"
)

// With returns a logger with the given key-value pairs attached,
// based on the logger extracted from the provided context.
func With(ctx context.Context, keyvals ...any) *log.Logger {
	return From(ctx).With(keyvals...)
}

// WithPrefix returns a logger with the given prefix attached,
// based on the logger extracted from the provided context.
func WithPrefix(ctx context.Context, prefix string) *log.Logger {
	return From(ctx).WithPrefix(prefix)
}

// WithTrace returns a logger with trace information (trace ID and span ID) attached,
// based on the logger extracted from the provided context and the current span from the context.
// If no valid span is present in the context, returns the logger without trace information.
func WithTrace(ctx context.Context) *log.Logger {
	logger := From(ctx)

	span := trace.SpanFromContext(ctx)
	if span == nil {
		return logger
	}

	if !span.SpanContext().IsValid() || !span.IsRecording() {
		return logger
	}

	return InjectTrace(logger, span)
}

// InjectTrace returns a logger with trace information (trace ID and span ID) attached.
func InjectTrace(logger *log.Logger, span trace.Span) *log.Logger {
	ctx := span.SpanContext()

	return logger.With(
		"trace_id", ctx.TraceID().String(),
		"span_id", ctx.SpanID().String(),
	)
}

// Debug logs a message at the debug level using the logger from the context.
func Debug(ctx context.Context, msg any, keyvals ...any) {
	From(ctx).Debug(msg, keyvals...)
}

// Info logs a message at the info level using the logger from the context.
func Info(ctx context.Context, msg any, keyvals ...any) {
	From(ctx).Info(msg, keyvals...)
}

// Warn logs a message at the warning level using the logger from the context.
func Warn(ctx context.Context, msg any, keyvals ...any) {
	From(ctx).Warn(msg, keyvals...)
}

// Error logs a message at the error level using the logger from the context.
func Error(ctx context.Context, msg any, keyvals ...any) {
	From(ctx).Error(msg, keyvals...)
}

// Debugf logs a formatted message at the debug level using the logger from the context.
func Debugf(ctx context.Context, format string, args ...any) {
	From(ctx).Debugf(format, args...)
}

// Infof logs a formatted message at the info level using the logger from the context.
func Infof(ctx context.Context, format string, args ...any) {
	From(ctx).Infof(format, args...)
}

// Warnf logs a formatted message at the warning level using the logger from the context.
func Warnf(ctx context.Context, format string, args ...any) {
	From(ctx).Warnf(format, args...)
}

// Errorf logs a formatted message at the error level using the logger from the context.
func Errorf(ctx context.Context, format string, args ...any) {
	From(ctx).Errorf(format, args...)
}
