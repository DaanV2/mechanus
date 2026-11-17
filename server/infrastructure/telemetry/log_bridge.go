package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/charmbracelet/log"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"
)

// LogSink returns a [logging.Sink] that emits every log line as an OpenTelemetry log record.
// Lines that carry the trace and span id (see [logging.From]) are correlated with that span.
func LogSink(logger otellog.Logger) logging.Sink {
	return func(entry *logging.Entry) {
		ctx, record := toRecord(entry)
		logger.Emit(ctx, record)
	}
}

func toRecord(entry *logging.Entry) (context.Context, otellog.Record) {
	var record otellog.Record
	record.SetTimestamp(entry.Time)
	record.SetObservedTimestamp(time.Now())
	record.SetBody(attribute.StringValue(entry.Message))
	record.SetSeverity(levelToSeverity(entry.Level))
	record.SetSeverityText(entry.Level.String())

	if entry.Prefix != "" {
		record.AddAttributes(attribute.String("prefix", entry.Prefix))
	}
	if entry.Caller != "" {
		record.AddAttributes(attribute.String("caller", entry.Caller))
	}

	var traceID trace.TraceID
	var spanID trace.SpanID
	for _, attr := range entry.Attrs {
		switch attr.Key {
		case logging.TraceIDKey:
			if s, ok := attr.Value.(string); ok {
				traceID, _ = trace.TraceIDFromHex(s)
			}
		case logging.SpanIDKey:
			if s, ok := attr.Value.(string); ok {
				spanID, _ = trace.SpanIDFromHex(s)
			}
		default:
			record.AddAttributes(attribute.KeyValue{Key: attribute.Key(attr.Key), Value: toValue(attr.Value)})
		}
	}

	ctx := context.Background()
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	if sc.IsValid() {
		ctx = trace.ContextWithSpanContext(ctx, sc)
	}

	return ctx, record
}

// toValue converts a JSON decoded log value to an OpenTelemetry value.
func toValue(value any) attribute.Value {
	switch v := value.(type) {
	case string:
		return attribute.StringValue(v)
	case bool:
		return attribute.BoolValue(v)
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return attribute.Int64Value(i)
		}
		if f, err := v.Float64(); err == nil {
			return attribute.Float64Value(f)
		}

		return attribute.StringValue(v.String())
	case nil:
		return attribute.Value{}
	case []any:
		values := make([]attribute.Value, 0, len(v))
		for _, item := range v {
			values = append(values, toValue(item))
		}

		return attribute.SliceValue(values...)
	case map[string]any:
		kvs := make([]attribute.KeyValue, 0, len(v))
		for key, item := range v {
			kvs = append(kvs, attribute.KeyValue{Key: attribute.Key(key), Value: toValue(item)})
		}

		return attribute.MapValue(kvs...)
	default:
		return attribute.StringValue(fmt.Sprint(v))
	}
}

func levelToSeverity(level log.Level) otellog.Severity {
	switch {
	case level >= log.FatalLevel:
		return otellog.SeverityFatal
	case level >= log.ErrorLevel:
		return otellog.SeverityError
	case level >= log.WarnLevel:
		return otellog.SeverityWarn
	case level >= log.InfoLevel:
		return otellog.SeverityInfo
	default:
		return otellog.SeverityDebug
	}
}
