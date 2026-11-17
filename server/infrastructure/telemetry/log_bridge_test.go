package telemetry_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/telemetry"
	"github.com/charmbracelet/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"
)

var _ = Describe("LogBridge", func() {
	Describe("ToRecord", func() {
		It("should convert a log entry into a record", func() {
			now := time.Now()
			ctx, record := telemetry.ToRecord(&logging.Entry{
				Time:    now,
				Level:   log.ErrorLevel,
				Prefix:  "scenes",
				Message: "failed to load",
				Caller:  "scenes/service.go:42",
				Attrs: []logging.Attr{
					{Key: "scene", Value: "tavern"},
					{Key: "attempt", Value: json.Number("2")},
					{Key: "ratio", Value: json.Number("0.5")},
					{Key: "ok", Value: false},
				},
			})

			Expect(trace.SpanContextFromContext(ctx).IsValid()).To(BeFalse())
			Expect(record.Timestamp()).To(Equal(now))
			Expect(record.Severity()).To(Equal(otellog.SeverityError))
			Expect(record.SeverityText()).To(Equal("error"))
			Expect(record.Body().AsString()).To(Equal("failed to load"))

			attrs := map[attribute.Key]attribute.Value{}
			record.WalkAttributes(func(kv attribute.KeyValue) bool {
				attrs[kv.Key] = kv.Value

				return true
			})
			Expect(attrs).To(HaveKeyWithValue(attribute.Key("prefix"), attribute.StringValue("scenes")))
			Expect(attrs).To(HaveKeyWithValue(attribute.Key("caller"), attribute.StringValue("scenes/service.go:42")))
			Expect(attrs).To(HaveKeyWithValue(attribute.Key("scene"), attribute.StringValue("tavern")))
			Expect(attrs).To(HaveKeyWithValue(attribute.Key("attempt"), attribute.Int64Value(2)))
			Expect(attrs).To(HaveKeyWithValue(attribute.Key("ratio"), attribute.Float64Value(0.5)))
			Expect(attrs).To(HaveKeyWithValue(attribute.Key("ok"), attribute.BoolValue(false)))
		})

		It("should correlate the record with the span of the logged ids", func() {
			ctx, record := telemetry.ToRecord(&logging.Entry{
				Level:   log.InfoLevel,
				Message: "traced",
				Attrs: []logging.Attr{
					{Key: logging.TraceIDKey, Value: "0102030405060708090a0b0c0d0e0f10"},
					{Key: logging.SpanIDKey, Value: "0102030405060708"},
				},
			})

			sc := trace.SpanContextFromContext(ctx)
			Expect(sc.IsValid()).To(BeTrue())
			Expect(sc.TraceID().String()).To(Equal("0102030405060708090a0b0c0d0e0f10"))
			Expect(sc.SpanID().String()).To(Equal("0102030405060708"))
			Expect(record.AttributesLen()).To(BeZero())
		})
	})
	Describe("Setup with logging enabled", func() {
		It("should initialize log provider when telemetry is enabled", func(ctx SpecContext) {
			// Stand-in collector: shutting down flushes the providers, which must not depend on a real one running
			collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			DeferCleanup(collector.Close)

			cfg := &telemetry.Config{
				Enabled:     true,
				Endpoint:    collector.Listener.Addr().String(),
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
