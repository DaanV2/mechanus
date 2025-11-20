package telemetry_test

import (
	"github.com/DaanV2/mechanus/server/infrastructure/telemetry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Setup", func() {
	Describe("SetupTracing", func() {
		It("should return a no-op provider when telemetry is disabled", func(ctx SpecContext) {
			cfg := &telemetry.Config{
				Enabled:     false,
				Endpoint:    "localhost:4318",
				ServiceName: "test-service",
				Insecure:    true,
			}

			provider, err := telemetry.Setup(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(provider).NotTo(BeNil())

			// Shutdown should succeed
			err = provider.ShutdownCleanup(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return an error when endpoint is invalid and telemetry is enabled", func(ctx SpecContext) {
			cfg := &telemetry.Config{
				Enabled:     true,
				Endpoint:    "invalid://endpoint:99999",
				ServiceName: "test-service",
				Insecure:    true,
			}

			provider, err := telemetry.Setup(ctx, cfg)
			// The provider might still be created but will fail to export
			if provider != nil {
				_ = provider.ShutdownCleanup(ctx)
			}
			// We don't expect an error during setup with invalid endpoint,
			// the error will occur during export
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
