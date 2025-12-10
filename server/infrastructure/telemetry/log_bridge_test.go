package telemetry_test

import (
	"bytes"
	"context"
	"log/slog" //nolint:depguard // Required for testing slog.Handler interface
	"strings"
	"time"

	"github.com/DaanV2/mechanus/server/infrastructure/telemetry"
	"github.com/charmbracelet/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LogBridge", func() {
	Describe("OtelLogHandler", func() {
		var (
			buf      *bytes.Buffer
			logger   *log.Logger
			handler  *telemetry.OtelLogHandler
			ctx      context.Context
		)

		BeforeEach(func() {
			buf = &bytes.Buffer{}
			logger = log.New(buf)
			handler = telemetry.NewOtelLogHandler(logger)
			ctx = context.Background()
		})

		It("should forward logs to the original handler", func() {
			record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
			record.AddAttrs(slog.String("key", "value"))

			err := handler.Handle(ctx, record)
			Expect(err).ToNot(HaveOccurred())

			// Verify the log was written to the original handler
			output := buf.String()
			Expect(output).To(ContainSubstring("test message"))
		})

		It("should respect the enabled level of the original handler", func() {
			logger.SetLevel(log.WarnLevel)

			enabled := handler.Enabled(ctx, slog.LevelInfo)
			Expect(enabled).To(BeFalse())

			enabled = handler.Enabled(ctx, slog.LevelWarn)
			Expect(enabled).To(BeTrue())
		})

		It("should support WithAttrs", func() {
			attrs := []slog.Attr{slog.String("service", "test")}
			newHandler := handler.WithAttrs(attrs)

			Expect(newHandler).ToNot(BeNil())
			Expect(newHandler).To(BeAssignableToTypeOf(&telemetry.OtelLogHandler{}))
		})

		It("should support WithGroup", func() {
			newHandler := handler.WithGroup("test-group")

			Expect(newHandler).ToNot(BeNil())
			Expect(newHandler).To(BeAssignableToTypeOf(&telemetry.OtelLogHandler{}))
		})
	})

	Describe("WrapLoggerWithOtel", func() {
		It("should set up the log bridge without panicking", func() {
			buf := &bytes.Buffer{}
			logger := log.New(buf)

			// This should not panic
			Expect(func() {
				telemetry.WrapLoggerWithOtel(logger)
			}).ToNot(Panic())

			// After wrapping, slog should still work
			slog.Info("test message after wrapping")
		})
	})

	Describe("Setup with logging enabled", func() {
		It("should initialize log provider when telemetry is enabled", func(ctx SpecContext) {
			cfg := &telemetry.Config{
				Enabled:     true,
				Endpoint:    "localhost:4318",
				ServiceName: "test-service",
				Insecure:    true,
			}

			manager, err := telemetry.Setup(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(manager).ToNot(BeNil())

			// Clean up
			err = manager.AfterShutDown(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not initialize log provider when telemetry is disabled", func(ctx SpecContext) {
			cfg := &telemetry.Config{
				Enabled:     false,
				Endpoint:    "localhost:4318",
				ServiceName: "test-service",
				Insecure:    true,
			}

			manager, err := telemetry.Setup(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(manager).ToNot(BeNil())

			// Clean up
			err = manager.AfterShutDown(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle log exporter initialization failure gracefully", func(ctx SpecContext) {
			buf := &bytes.Buffer{}
			oldLogger := log.Default()
			defer log.SetDefault(oldLogger)

			// Set up a logger that captures warnings
			testLogger := log.New(buf)
			log.SetDefault(testLogger)

			cfg := &telemetry.Config{
				Enabled:     true,
				Endpoint:    "invalid://endpoint:99999",
				ServiceName: "test-service",
				Insecure:    true,
			}

			manager, err := telemetry.Setup(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(manager).ToNot(BeNil())

			// Verify that a warning was logged about log exporter failure
			output := buf.String()
			Expect(strings.Contains(output, "failed to create OTLP log exporter") || 
			       strings.Contains(output, "WARN")).To(BeTrue())

			// Clean up
			err = manager.AfterShutDown(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
