package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/resource"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// SetupLogging initializes OpenTelemetry logging with the given configuration
func SetupLogging(ctx context.Context, cfg *Config) (*sdklog.LoggerProvider, error) {
	if !cfg.Enabled {
		// Return a no-op logger provider when logging is disabled
		return sdklog.NewLoggerProvider(), nil
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

	// Create OTLP HTTP log exporter
	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(cfg.Endpoint),
	}

	if cfg.Insecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}

	exporter, err := otlploghttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}

	// Create logger provider with batch processor
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)

	// Set global logger provider
	global.SetLoggerProvider(lp)

	return lp, nil
}

// ShutdownLogging gracefully shuts down the logger provider
func ShutdownLogging(ctx context.Context, lp *sdklog.LoggerProvider) error {
	if lp == nil {
		return nil
	}

	return lp.Shutdown(ctx)
}
