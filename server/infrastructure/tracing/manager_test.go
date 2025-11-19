package tracing_test

import (
	"github.com/DaanV2/mechanus/server/infrastructure/tracing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var _ = Describe("Manager", func() {
	Describe("Lifecycle", func() {
		It("should handle lifecycle with no-op provider", func(ctx SpecContext) {
			provider := sdktrace.NewTracerProvider()
			logProvider := sdklog.NewLoggerProvider()
			manager := tracing.NewManager(provider, logProvider)

			err := manager.AfterInitialize(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = manager.BeforeShutdown(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = manager.AfterShutDown(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = manager.ShutdownCleanup(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle lifecycle with nil provider", func(ctx SpecContext) {
			manager := tracing.NewManager(nil, nil)

			err := manager.AfterInitialize(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = manager.BeforeShutdown(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = manager.AfterShutDown(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = manager.ShutdownCleanup(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
