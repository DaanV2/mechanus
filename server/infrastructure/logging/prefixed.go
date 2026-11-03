package logging

import (
	"context"

	"github.com/charmbracelet/log"
)

type Prefixed string

func (e Prefixed) With(keyvalues ...any) Enriched {
	return Enriched{
		prefix: string(e),
		values: keyvalues,
	}
}

// Example:
//
//	logger := p.From(ctx, "userId", jwt.Claims.User.ID, "campaignName", name)
func (p Prefixed) From(ctx context.Context, keyvalues ...any) *log.Logger {
	l := From(ctx).WithPrefix(string(p))
	if len(keyvalues) > 0 {
		l = l.With(keyvalues...)
	}

	return l
}

func (p Prefixed) FromUpdate(ctx context.Context) (*log.Logger, context.Context) {
	l := p.From(ctx)

	return l, Context(ctx, l)
}

// Debug prints a debug message.
func (l Prefixed) Debug(ctx context.Context, msg any, keyvals ...any) {
	l.From(ctx).Debug(msg, keyvals...)
}

// Info prints an info message.
func (l Prefixed) Info(ctx context.Context, msg any, keyvals ...any) {
	l.From(ctx).Info(msg, keyvals...)
}

// Warn prints a warning message.
func (l Prefixed) Warn(ctx context.Context, msg any, keyvals ...any) {
	l.From(ctx).Warn(msg, keyvals...)
}

// Error prints an error message.
func (l Prefixed) Error(ctx context.Context, msg any, keyvals ...any) {
	l.From(ctx).Error(msg, keyvals...)
}

// Fatal prints a fatal message and exits.
func (l Prefixed) Fatal(ctx context.Context, msg any, keyvals ...any) {
	l.From(ctx).Fatal(msg, keyvals...)
}

// Print prints a message with no level.
func (l Prefixed) Print(ctx context.Context, msg any, keyvals ...any) {
	l.From(ctx).Print(msg, keyvals...)
}

// Debugf prints a debug message with formatting.
func (l Prefixed) Debugf(ctx context.Context, format string, args ...any) {
	l.From(ctx).Debugf(format, args...)
}

// Infof prints an info message with formatting.
func (l Prefixed) Infof(ctx context.Context, format string, args ...any) {
	l.From(ctx).Infof(format, args...)
}

// Warnf prints a warning message with formatting.
func (l Prefixed) Warnf(ctx context.Context, format string, args ...any) {
	l.From(ctx).Warnf(format, args...)
}

// Errorf prints an error message with formatting.
func (l Prefixed) Errorf(ctx context.Context, format string, args ...any) {
	l.From(ctx).Errorf(format, args...)
}

// Fatalf prints a fatal message with formatting and exits.
func (l Prefixed) Fatalf(ctx context.Context, format string, args ...any) {
	l.From(ctx).Fatalf(format, args...)
}

// Printf prints a message with no level and formatting.
func (l Prefixed) Printf(ctx context.Context, format string, args ...any) {
	l.From(ctx).Printf(format, args...)
}
