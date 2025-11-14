package tracing_test

import (
	"context"

	"github.com/DaanV2/mechanus/server/infrastructure/tracing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var _ = Describe("Manager", func() {
	ctx := context.Background()

	Describe("Lifecycle", func() {
		It("should handle lifecycle with no-op provider", func() {
			provider := sdktrace.NewTracerProvider()
			manager := tracing.NewManager(provider)

			err := manager.AfterInitialize(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = manager.BeforeShutdown(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = manager.AfterShutDown(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = manager.ShutdownCleanup(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle lifecycle with nil provider", func() {
			manager := tracing.NewManager(nil)

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
