package lifecycle

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Lifecycle calls run outside of any request, these spans give the work they do (database queries, ...) a parent.
// AfterShutDown and ShutdownCleanup aren't traced: they shut down the tracer provider itself.
var tracer = otel.Tracer("github.com/DaanV2/mechanus/server/infrastructure/lifecycle")

// traced runs call in a span named after the component.
func traced(ctx context.Context, component any, call func(ctx context.Context) error) error {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("%T", component))
	defer span.End()

	return endSpan(span, call(ctx))
}

// endSpan marks the span as failed when err is set, and returns err.
func endSpan(span trace.Span, err error) error {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}
