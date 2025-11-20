package telemetry

import (
	"context"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/charmbracelet/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// SpanLogger starts a new span and returns a context containing the span and a SpanLogger
func SpanLogger(ctx context.Context, name string, attributes ...attribute.KeyValue) (context.Context, trace.Span, *log.Logger) {
	ctx, span := otel.Tracer("default").Start(ctx, name)
	logger := logging.From(ctx)

	if len(attributes) > 0 {
		span.SetAttributes(attributes...)
		logger = logger.With(otelToCharmlogAttrs(attributes))
	}
	logger = logging.InjectTrace(logger, span)
	ctx = logging.Context(ctx, logger)

	return ctx, span, logger
}
