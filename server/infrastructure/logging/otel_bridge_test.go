package logging_test

import (
	"context"
	// nolint:depguard // this is needed for testing
	"log/slog"
	"testing"
	"time"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

func TestLogging(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logging Suite")
}

var _ = Describe("OtelBridge", func() {
	var (
		bridge   *logging.OtelBridge
		mockBase *mockHandler
		exporter *mockExporter
		provider *sdklog.LoggerProvider
		ctx      context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockBase = &mockHandler{enabled: true}

		// Set up OTEL log provider with mock exporter
		exporter = &mockExporter{}
		provider = sdklog.NewLoggerProvider(
			sdklog.WithProcessor(sdklog.NewSimpleProcessor(exporter)),
		)
		global.SetLoggerProvider(provider)

		bridge = logging.NewOtelBridge(mockBase)
	})

	AfterEach(func() {
		if provider != nil {
			provider.Shutdown(ctx)
		}
	})

	Describe("Handle", func() {
		It("should forward logs to both base handler and OTEL", func() {
			record := slog.Record{
				Time:    time.Now(),
				Message: "test message",
				Level:   slog.LevelInfo,
			}
			record.AddAttrs(slog.String("key", "value"))

			err := bridge.Handle(ctx, record)
			Expect(err).ToNot(HaveOccurred())

			// Verify base handler was called
			Expect(mockBase.called).To(BeTrue())
			Expect(mockBase.lastRecord.Message).To(Equal("test message"))

			// Verify OTEL received the log (processor needs to be flushed)
			Eventually(func() int {
				return len(exporter.records)
			}, "1s").Should(BeNumerically(">", 0))
		})

		It("should handle multiple log levels", func() {
			testCases := []struct {
				level   slog.Level
				message string
			}{
				{slog.LevelDebug, "debug message"},
				{slog.LevelInfo, "info message"},
				{slog.LevelWarn, "warn message"},
				{slog.LevelError, "error message"},
			}

			for _, tc := range testCases {
				record := slog.Record{
					Time:    time.Now(),
					Message: tc.message,
					Level:   tc.level,
				}

				err := bridge.Handle(ctx, record)
				Expect(err).ToNot(HaveOccurred())
			}

			// Verify base handler was called multiple times
			Expect(mockBase.called).To(BeTrue())
		})
	})

	Describe("Enabled", func() {
		It("should delegate to base handler", func() {
			mockBase.enabled = true
			Expect(bridge.Enabled(ctx, slog.LevelInfo)).To(BeTrue())

			mockBase.enabled = false
			Expect(bridge.Enabled(ctx, slog.LevelInfo)).To(BeFalse())
		})
	})

	Describe("WithAttrs", func() {
		It("should create a new bridge with attributes", func() {
			attrs := []slog.Attr{slog.String("key", "value")}
			newBridge := bridge.WithAttrs(attrs)
			Expect(newBridge).ToNot(BeNil())
		})
	})

	Describe("WithGroup", func() {
		It("should create a new bridge with group", func() {
			newBridge := bridge.WithGroup("test-group")
			Expect(newBridge).ToNot(BeNil())
		})
	})
})

// mockHandler is a simple mock implementation of slog.Handler for testing
type mockHandler struct {
	called     bool
	lastRecord slog.Record
	enabled    bool
}

func (m *mockHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return m.enabled
}

func (m *mockHandler) Handle(_ context.Context, record slog.Record) error {
	m.called = true
	m.lastRecord = record
	return nil
}

func (m *mockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return m
}

func (m *mockHandler) WithGroup(name string) slog.Handler {
	return m
}

// mockExporter is a simple mock implementation of sdklog.Exporter for testing
type mockExporter struct {
	records []sdklog.Record
}

func (m *mockExporter) Export(ctx context.Context, records []sdklog.Record) error {
	m.records = append(m.records, records...)
	return nil
}

func (m *mockExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (m *mockExporter) ForceFlush(ctx context.Context) error {
	return nil
}
