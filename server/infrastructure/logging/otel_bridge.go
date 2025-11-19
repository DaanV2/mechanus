package logging

import (
	"context"
	// nolint:depguard // this is needed for otel log bridge
	"log/slog"

	"github.com/charmbracelet/log"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
)

// OtelBridge is a slog.Handler that forwards logs to OpenTelemetry
type OtelBridge struct {
	base   slog.Handler
	logger otellog.Logger
}

// NewOtelBridge creates a new bridge handler that forwards logs to both
// the base handler and OpenTelemetry
func NewOtelBridge(base slog.Handler) *OtelBridge {
	return &OtelBridge{
		base:   base,
		logger: global.GetLoggerProvider().Logger("mechanus"),
	}
}

// Enabled reports whether the handler handles records at the given level
func (h *OtelBridge) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

// Handle handles the log record, forwarding it to both the base handler and OTEL
func (h *OtelBridge) Handle(ctx context.Context, record slog.Record) error {
	// Forward to base handler (charm logger)
	if err := h.base.Handle(ctx, record); err != nil {
		return err
	}

	// Forward to OTEL
	if h.logger != nil {
		h.emitToOtel(ctx, record)
	}

	return nil
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments
func (h *OtelBridge) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &OtelBridge{
		base:   h.base.WithAttrs(attrs),
		logger: h.logger,
	}
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups
func (h *OtelBridge) WithGroup(name string) slog.Handler {
	return &OtelBridge{
		base:   h.base.WithGroup(name),
		logger: h.logger,
	}
}

// emitToOtel converts slog.Record to OTEL log and emits it
func (h *OtelBridge) emitToOtel(ctx context.Context, record slog.Record) {
	// Convert slog level to OTEL severity
	severity := slogLevelToOtelSeverity(record.Level)

	// Build attributes
	var attrs []otellog.KeyValue
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, otellog.String(attr.Key, attr.Value.String()))
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
	h.logger.Emit(ctx, logRecord)
}

// slogLevelToOtelSeverity converts slog.Level to OTEL log severity
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

// SetupOtelBridge wraps the default charm logger with an OTEL bridge
// This should be called after the OTEL log provider is set up
func SetupOtelBridge() {
	defaultLogger := log.Default()

	// Create a bridge handler that wraps the charm logger
	bridge := NewOtelBridge(defaultLogger)

	// Create a new slog logger with the bridge handler
	slogLogger := slog.New(bridge)

	// Set as default slog logger
	slog.SetDefault(slogLogger)
}
