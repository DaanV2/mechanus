package telemetry

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Setup initializes OpenTelemetry telemetry with the given configuration
func Setup(ctx context.Context, cfg *Config) (*Manager, error) {
	otel.SetErrorHandler(&otelErrorHandler{})
	manager := NewManager()

	if !cfg.Enabled {
		return manager, nil
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP HTTP trace exporter
	traceOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		traceOpts = append(traceOpts, otlptracehttp.WithInsecure())
	}

	traceExporter, err := otlptracehttp.New(ctx, traceOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	manager.SetTraceProvider(SetupTracing(traceExporter, res))
	manager.SetExporter(traceExporter)

	// Create OTLP HTTP log exporter
	logOpts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		logOpts = append(logOpts, otlploghttp.WithInsecure())
	}

	logExporter, err := otlploghttp.New(ctx, logOpts...)
	if err != nil {
		// Log the error but don't fail setup - the log exporter will fail during export
		log.Warn("failed to create OTLP log exporter", "error", err)
	} else {
		manager.SetLogProvider(SetupLogging(logExporter, res))
		manager.SetLogExporter(logExporter)

		// Setup the log bridge to forward charm.sh logs to OpenTelemetry
		WrapLoggerWithOtel(log.Default())
	}

	return manager, nil
}

func SetupTracing(exporter *otlptrace.Exporter, res *resource.Resource) *sdktrace.TracerProvider {
	// Create tracer provider with batch span processor
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator for distributed telemetry
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp
}

// SetupLogging creates and configures the OpenTelemetry log provider
func SetupLogging(exporter *otlploghttp.Exporter, res *resource.Resource) *sdklog.LoggerProvider {
	// Create logger provider with batch processor
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)

	// Set global logger provider
	global.SetLoggerProvider(lp)

	return lp
}