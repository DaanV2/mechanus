package telemetry

import (
	"context"
	"log/slog" //nolint:depguard // Required for slog.Handler interface compatibility
	"time"

	"github.com/charmbracelet/log"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
)

// OtelLogHandler is a slog.Handler that forwards log entries to OpenTelemetry.
// It wraps the original charm.sh logger handler to maintain existing functionality
// while also sending logs to the OpenTelemetry collector.
type OtelLogHandler struct {
	original slog.Handler
}

// NewOtelLogHandler creates a new handler that forwards logs to both
// the original handler and OpenTelemetry.
func NewOtelLogHandler(original slog.Handler) *OtelLogHandler {
	return &OtelLogHandler{
		original: original,
	}
}

// Enabled returns true if the handler is enabled for the given level.
func (h *OtelLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.original.Enabled(ctx, level)
}

// Handle forwards the log record to both the original handler and OpenTelemetry.
func (h *OtelLogHandler) Handle(ctx context.Context, record slog.Record) error { //nolint:gocritic // slog.Record is passed by value by design
	// First, handle with the original handler to maintain existing behavior
	if err := h.original.Handle(ctx, record); err != nil {
		return err
	}

	// Convert and send to OpenTelemetry
	h.sendToOtel(ctx, record)
	
	return nil
}

// WithAttrs returns a new handler with the given attributes added.
func (h *OtelLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &OtelLogHandler{
		original: h.original.WithAttrs(attrs),
	}
}

// WithGroup returns a new handler with the given group added.
func (h *OtelLogHandler) WithGroup(name string) slog.Handler {
	return &OtelLogHandler{
		original: h.original.WithGroup(name),
	}
}

// sendToOtel converts the slog.Record to an OpenTelemetry log record and emits it.
func (h *OtelLogHandler) sendToOtel(ctx context.Context, record slog.Record) { //nolint:gocritic // slog.Record is passed by value by design
	logger := global.GetLoggerProvider().Logger("charm.sh")
	
	// Convert slog level to OpenTelemetry severity
	severity := slogLevelToOtelSeverity(record.Level)
	
	// Collect attributes
	var attrs []otellog.KeyValue
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, slogAttrToOtelKeyValue(attr))

		return true
	})
	
	// Create log record
	var logRecord otellog.Record
	logRecord.SetTimestamp(record.Time)
	logRecord.SetBody(otellog.StringValue(record.Message))
	logRecord.SetSeverity(severity)
	logRecord.SetSeverityText(record.Level.String())
	logRecord.AddAttributes(attrs...)
	
	// Emit the log record
	logger.Emit(ctx, logRecord)
}

// slogLevelToOtelSeverity converts a slog.Level to OpenTelemetry severity.
func slogLevelToOtelSeverity(level slog.Level) otellog.Severity {
	switch {
	case level >= slog.LevelError:
		return otellog.SeverityError
	case level >= slog.LevelWarn:
		return otellog.SeverityWarn
	case level >= slog.LevelInfo:
		return otellog.SeverityInfo
	default:
		return otellog.SeverityDebug
	}
}

// slogAttrToOtelKeyValue converts a slog.Attr to an OpenTelemetry KeyValue.
func slogAttrToOtelKeyValue(attr slog.Attr) otellog.KeyValue {
	key := attr.Key
	value := attr.Value
	
	switch value.Kind() {
	case slog.KindBool:
		return otellog.Bool(key, value.Bool())
	case slog.KindInt64:
		return otellog.Int64(key, value.Int64())
	case slog.KindUint64:
		// Convert uint64 to int64, clamping at max int64 if needed
		u := value.Uint64()
		if u > 9223372036854775807 { // max int64
			return otellog.Int64(key, 9223372036854775807)
		}

		return otellog.Int64(key, int64(u)) //nolint:gosec // Checked for overflow above
	case slog.KindFloat64:
		return otellog.Float64(key, value.Float64())
	case slog.KindString:
		return otellog.String(key, value.String())
	case slog.KindTime:
		return otellog.String(key, value.Time().Format(time.RFC3339))
	case slog.KindDuration:
		return otellog.Int64(key, value.Duration().Milliseconds())
	case slog.KindGroup:
		// For groups, we could flatten them or represent as a string
		return otellog.String(key, value.String())
	case slog.KindLogValuer:
		return otellog.String(key, value.String())
	case slog.KindAny:
		return otellog.String(key, value.String())
	default:
		return otellog.String(key, value.String())
	}
}

// WrapLoggerWithOtel wraps the charm.sh logger with an OpenTelemetry bridge.
// This intercepts log calls and forwards them to OpenTelemetry while maintaining
// the original logging behavior. Since charm.sh logger implements slog.Handler,
// we wrap it and set it as the default slog handler.
func WrapLoggerWithOtel(logger *log.Logger) {
	// Wrap the logger (which implements slog.Handler) with our OpenTelemetry handler
	handler := NewOtelLogHandler(logger)
	
	// Set the wrapped handler as the default slog handler
	// This ensures all slog-based logging goes through our bridge
	slog.SetDefault(slog.New(handler))
}
