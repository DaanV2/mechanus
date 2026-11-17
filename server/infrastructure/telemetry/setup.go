package telemetry

import (
	"context"
	"fmt"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/mechanus"
	"github.com/charmbracelet/log"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
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
		lp := SetupLogging(logExporter, res)
		manager.SetLogProvider(lp)
		manager.SetLogExporter(logExporter)

		// Forward everything logged through charm.sh to OpenTelemetry
		logging.SetSink(LogSink(lp.Logger(mechanus.SERVICE_NAME))) //nolint:contextcheck // log lines arrive without a context, the span is rebuilt from the logged ids
	}

	// Create OTLP HTTP metric exporter
	metricOpts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		metricOpts = append(metricOpts, otlpmetrichttp.WithInsecure())
	}

	metricExporter, err := otlpmetrichttp.New(ctx, metricOpts...)
	if err != nil {
		// Like the log exporter, don't fail setup over it - traces may still work
		log.Warn("failed to create OTLP metric exporter", "error", err)

		return manager, nil
	}

	mp := SetupMetrics(metricExporter, res)
	manager.SetMeterProvider(mp)

	// Go runtime metrics: memory, GC, goroutines, ...
	if err := runtime.Start(runtime.WithMeterProvider(mp)); err != nil {
		return nil, fmt.Errorf("failed to start runtime metrics: %w", err)
	}

	return manager, nil
}

// SetupMetrics creates and configures the OpenTelemetry meter provider.
// The HTTP, RPC and database instrumentation picks it up through the global meter provider.
func SetupMetrics(exporter sdkmetric.Exporter, res *resource.Resource) *sdkmetric.MeterProvider {
	// The export interval can be tuned with OTEL_METRIC_EXPORT_INTERVAL (milliseconds, default 60s)
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
		sdkmetric.WithResource(res),
	)

	// Set global meter provider
	otel.SetMeterProvider(mp)

	return mp
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
	otel.SetLoggerProvider(lp)

	return lp
}
