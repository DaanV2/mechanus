package lifecycle_test

import (
	"context"
	"errors"

	"github.com/DaanV2/mechanus/server/infrastructure/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

type initializer struct {
	err    error
	parent trace.SpanContext
}

func (i *initializer) AfterInitialize(ctx context.Context) error {
	i.parent = trace.SpanContextFromContext(ctx)

	return i.err
}

var _ = Describe("Tracing", func() {
	var exporter *tracetest.InMemoryExporter

	BeforeEach(func() {
		exporter = tracetest.NewInMemoryExporter()
		previous := otel.GetTracerProvider()
		otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter)))
		DeferCleanup(func() { otel.SetTracerProvider(previous) })
	})

	It("should run every component in its own span, under a span of the phase", func(ctx SpecContext) {
		component := &initializer{}
		manager := lifecycle.NewManager()
		manager.Add(component)

		Expect(manager.AfterInitialize(ctx)).To(Succeed())

		spans := exporter.GetSpans()
		Expect(spans).To(HaveLen(2))
		child, phase := spans[0], spans[1]
		Expect(phase.Name).To(Equal("lifecycle.AfterInitialize"))
		Expect(child.Name).To(Equal("*lifecycle_test.initializer"))
		Expect(child.Parent.SpanID()).To(Equal(phase.SpanContext.SpanID()))
		// The component gets the context of its own span, so its work nests under it
		Expect(component.parent.SpanID()).To(Equal(child.SpanContext.SpanID()))
	})

	It("should mark the spans of a failing component", func(ctx SpecContext) {
		manager := lifecycle.NewManager()
		manager.Add(&initializer{err: errors.New("boom")})

		Expect(manager.AfterInitialize(ctx)).ToNot(Succeed())

		for _, span := range exporter.GetSpans() {
			Expect(span.Status.Code).To(Equal(codes.Error))
		}
	})
})
