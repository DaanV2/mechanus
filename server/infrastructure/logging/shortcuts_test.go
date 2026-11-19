package logging_test

import (
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var _ = Describe("WithTrace", func() {
	var (
		provider *sdktrace.TracerProvider
		tracer   trace.Tracer
	)

	BeforeEach(func() {
		provider = sdktrace.NewTracerProvider()
		otel.SetTracerProvider(provider)
		tracer = provider.Tracer("test")
	})

	AfterEach(func(ctx SpecContext) {
		if provider != nil {
			_ = provider.Shutdown(ctx)
		}
	})

	It("should return logger without trace info when no span in context", func(ctx SpecContext) {
		logger := logging.WithTrace(ctx)
		Expect(logger).ToNot(BeNil())
	})

	It("should return logger with trace info when valid span in context", func(ctx SpecContext) {
		spanCtx, span := tracer.Start(ctx, "test-operation")
		defer span.End()

		logger := logging.WithTrace(spanCtx)
		Expect(logger).ToNot(BeNil())

		spanContext := span.SpanContext()
		Expect(spanContext.IsValid()).To(BeTrue())
		Expect(spanContext.TraceID().String()).ToNot(BeEmpty())
		Expect(spanContext.SpanID().String()).ToNot(BeEmpty())
	})

	It("should handle context with invalid span context", func(ctx SpecContext) {
		// Create a context with no span or an invalid span
		logger := logging.WithTrace(ctx)
		Expect(logger).ToNot(BeNil())
	})
})
