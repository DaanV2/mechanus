package tracing_test

import (
	"context"

	"github.com/DaanV2/mechanus/server/infrastructure/tracing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupLogging", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("with disabled config", func() {
		It("should return a no-op logger provider", func() {
			cfg := &tracing.Config{
				Enabled: false,
			}

			provider, err := tracing.SetupLogging(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(provider).ToNot(BeNil())

			// Cleanup
			err = tracing.ShutdownLogging(ctx, provider)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("with enabled config", func() {
		It("should create a logger provider with exporter", func() {
			cfg := &tracing.Config{
				Enabled:     true,
				Endpoint:    "localhost:4318",
				ServiceName: "test-service",
				Insecure:    true,
			}

			provider, err := tracing.SetupLogging(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(provider).ToNot(BeNil())

			// Cleanup
			err = tracing.ShutdownLogging(ctx, provider)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("ShutdownLogging", func() {
		It("should handle nil provider gracefully", func() {
			err := tracing.ShutdownLogging(ctx, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should shutdown provider without error", func() {
			cfg := &tracing.Config{
				Enabled:     true,
				Endpoint:    "localhost:4318",
				ServiceName: "test-service",
				Insecure:    true,
			}

			provider, err := tracing.SetupLogging(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = tracing.ShutdownLogging(ctx, provider)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
