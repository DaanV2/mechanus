package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/charmbracelet/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/trace"
)

var _ = Describe("Bridge", func() {
	var (
		console *bytes.Buffer
		entries []logging.Entry
		mu      sync.Mutex
	)

	BeforeEach(func() {
		oldLogger := log.Default()
		DeferCleanup(func() {
			log.SetDefault(oldLogger)
			logging.ResetSink()
		})

		console = &bytes.Buffer{}
		entries = nil
		logging.UseConsole(console)
		log.SetDefault(log.NewWithOptions(console, log.Options{Level: log.DebugLevel}))

		logging.SetSink(func(entry *logging.Entry) {
			mu.Lock()
			defer mu.Unlock()
			entries = append(entries, *entry)
		})
	})

	It("should hand every line to the sink as a structured entry", func() {
		log.Default().WithPrefix("scenes").Warn("scene not found", "scene", "tavern", "attempt", 2, "ok", false)

		Expect(entries).To(HaveLen(1))
		entry := entries[0]
		Expect(entry.Level).To(Equal(log.WarnLevel))
		Expect(entry.Prefix).To(Equal("scenes"))
		Expect(entry.Message).To(Equal("scene not found"))
		Expect(entry.Time).ToNot(BeZero())
		Expect(entry.Attrs).To(Equal([]logging.Attr{
			{Key: "scene", Value: "tavern"},
			{Key: "attempt", Value: json.Number("2")},
			{Key: "ok", Value: false},
		}))
	})

	It("should keep printing to the console in the configured format", func() {
		logging.ConsoleFormatter(log.LogfmtFormatter)
		log.Default().Info("hello", "error", errors.New("boom"))

		Expect(console.String()).To(ContainSubstring(`level=info msg=hello error=boom`))
	})

	It("should respect the level of the default logger", func() {
		log.Default().SetLevel(log.InfoLevel)
		log.Default().Debug("hidden")

		Expect(entries).To(BeEmpty())
		Expect(console.String()).To(BeEmpty())
	})

	It("should add the ids of the active span, without printing them", func() {
		traceID, err := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
		Expect(err).ToNot(HaveOccurred())
		spanID, err := trace.SpanIDFromHex("0102030405060708")
		Expect(err).ToNot(HaveOccurred())
		ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: traceID,
			SpanID:  spanID,
		}))

		logging.From(ctx).Info("traced")

		Expect(entries).To(HaveLen(1))
		Expect(entries[0].Attrs).To(ContainElements(
			logging.Attr{Key: logging.TraceIDKey, Value: traceID.String()},
			logging.Attr{Key: logging.SpanIDKey, Value: spanID.String()},
		))
		Expect(console.String()).To(ContainSubstring("traced"))
		Expect(console.String()).ToNot(ContainSubstring(traceID.String()))
	})

	It("should not add span ids without an active span", func() {
		logging.From(context.Background()).Info("untraced")

		Expect(entries).To(HaveLen(1))
		Expect(entries[0].Attrs).To(BeEmpty())
	})
})
