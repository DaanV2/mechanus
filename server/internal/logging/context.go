package logging

import (
	"context"

	"github.com/charmbracelet/log"
)

type loggerContextKey struct{}

// From returns the possible injected logger in the context
func From(ctx context.Context) *log.Logger {
	v := ctx.Value(loggerContextKey{})
	if v != nil {
		logger, ok := v.(*log.Logger)
		if ok {
			return logger
		}
	}

	return log.Default()
}

// Context updates the given logger with the updated logger
func Context(ctx context.Context, logger *log.Logger) context.Context {
	if logger == nil {
		return ctx
	}

	return context.WithValue(ctx, loggerContextKey{}, logger)
}
