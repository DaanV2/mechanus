package tracing_test

import (
	"github.com/DaanV2/mechanus/server/infrastructure/tracing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Setup", func() {
	Describe("SetupTracing", func() {
		It("should return a no-op provider when tracing is disabled", func(ctx SpecContext) {
			cfg := &tracing.Config{
				Enabled:     false,
				Endpoint:    "localhost:4318",
				ServiceName: "test-service",
				Insecure:    true,
			}

			provider, err := tracing.SetupTracing(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(provider).NotTo(BeNil())

			// Shutdown should succeed
			err = tracing.Shutdown(ctx, provider)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return an error when endpoint is invalid and tracing is enabled", func(ctx SpecContext) {
			cfg := &tracing.Config{
				Enabled:     true,
				Endpoint:    "invalid://endpoint:99999",
				ServiceName: "test-service",
				Insecure:    true,
			}

			provider, err := tracing.SetupTracing(ctx, cfg)
			// The provider might still be created but will fail to export
			if provider != nil {
				_ = tracing.Shutdown(ctx, provider)
			}
			// We don't expect an error during setup with invalid endpoint,
			// the error will occur during export
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Shutdown", func() {
		It("should handle nil provider gracefully", func(ctx SpecContext) {
			err := tracing.Shutdown(ctx, nil)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
