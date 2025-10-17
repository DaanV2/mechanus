package logging

import (
	"context"

	"github.com/charmbracelet/log"
)

type Enriched struct {
	prefix string
	values []any
}

func (e Enriched) WithPrefix(prefix string) Enriched {
	e.prefix = prefix

	return e
}

func (e Enriched) With(keyvalues ...interface{}) Enriched {
	e.values = append(e.values, keyvalues...)

	return e
}

func (p Enriched) From(ctx context.Context) *log.Logger {
	l := From(ctx)
	if p.prefix != "" {
		l = l.WithPrefix(p.prefix)
	}

	if len(p.values) > 0 {
		l = l.With(p.values...)
	}

	return l
}

func (p Enriched) FromUpdate(ctx context.Context) (*log.Logger, context.Context) {
	l := p.From(ctx)

	return l, Context(ctx, l)
}

// Debug prints a debug message.
func (l Enriched) Debug(ctx context.Context, msg interface{}, keyvals ...interface{}) {
	l.From(ctx).Debug(msg, keyvals...)
}

// Info prints an info message.
func (l Enriched) Info(ctx context.Context, msg interface{}, keyvals ...interface{}) {
	l.From(ctx).Info(msg, keyvals...)
}

// Warn prints a warning message.
func (l Enriched) Warn(ctx context.Context, msg interface{}, keyvals ...interface{}) {
	l.From(ctx).Warn(msg, keyvals...)
}

// Error prints an error message.
func (l Enriched) Error(ctx context.Context, msg interface{}, keyvals ...interface{}) {
	l.From(ctx).Error(msg, keyvals...)
}

// Fatal prints a fatal message and exits.
func (l Enriched) Fatal(ctx context.Context, msg interface{}, keyvals ...interface{}) {
	l.From(ctx).Fatal(msg, keyvals...)
}

// Print prints a message with no level.
func (l Enriched) Print(ctx context.Context, msg interface{}, keyvals ...interface{}) {
	l.From(ctx).Print(msg, keyvals...)
}

// Debugf prints a debug message with formatting.
func (l Enriched) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.From(ctx).Debugf(format, args...)
}

// Infof prints an info message with formatting.
func (l Enriched) Infof(ctx context.Context, format string, args ...interface{}) {
	l.From(ctx).Infof(format, args...)
}

// Warnf prints a warning message with formatting.
func (l Enriched) Warnf(ctx context.Context, format string, args ...interface{}) {
	l.From(ctx).Warnf(format, args...)
}

// Errorf prints an error message with formatting.
func (l Enriched) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.From(ctx).Errorf(format, args...)
}

// Fatalf prints a fatal message with formatting and exits.
func (l Enriched) Fatalf(ctx context.Context, format string, args ...interface{}) {
	l.From(ctx).Fatalf(format, args...)
}

// Printf prints a message with no level and formatting.
func (l Enriched) Printf(ctx context.Context, format string, args ...interface{}) {
	l.From(ctx).Printf(format, args...)
}
