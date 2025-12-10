package telemetry

import (
	"context"
	"errors"
	"fmt"

	"github.com/DaanV2/mechanus/server/infrastructure/lifecycle"
	"github.com/charmbracelet/log"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var (
	_ lifecycle.AfterInitialize = &Manager{}
	_ lifecycle.AfterShutDown = &Manager{}
	_ lifecycle.BeforeShutdown = &Manager{}
	_ lifecycle.ShutdownCleanup = &Manager{}
)

// Manager handles the lifecycle of the OpenTelemetry providers
type Manager struct {
	traceProvider *sdktrace.TracerProvider
	traceExporter *otlptrace.Exporter
	logProvider   *sdklog.LoggerProvider
	logExporter   *otlploghttp.Exporter
}

// NewManager creates a new telemetry manager
func NewManager() *Manager {
	return &Manager{
		traceProvider: nil,
		traceExporter: nil,
		logProvider:   nil,
		logExporter:   nil,
	}
}

func (m *Manager) SetTraceProvider(provider *sdktrace.TracerProvider) { m.traceProvider = provider }
func (m *Manager) SetExporter(exporter *otlptrace.Exporter)           { m.traceExporter = exporter }
func (m *Manager) SetLogProvider(provider *sdklog.LoggerProvider)     { m.logProvider = provider }
func (m *Manager) SetLogExporter(exporter *otlploghttp.Exporter)      { m.logExporter = exporter }

// AfterInitialize is called during the initialization phase
func (m *Manager) AfterInitialize(ctx context.Context) error {
	if m.traceProvider != nil {
		log.Debug("OpenTelemetry trace provider initialized")
	}
	if m.traceExporter != nil {
		log.Debug("OpenTelemetry trace exporter initialized")
	}
	if m.logProvider != nil {
		log.Debug("OpenTelemetry log provider initialized")
	}
	if m.logExporter != nil {
		log.Debug("OpenTelemetry log exporter initialized")
	}

	return nil
}

// BeforeShutdown is called before shutdown begins
func (m *Manager) BeforeShutdown(ctx context.Context) error {
	return nil
}

// AfterShutDown is called after shutdown completes
func (m *Manager) AfterShutDown(ctx context.Context) error {
	var err error

	if m.traceProvider != nil {
		log.Debug("Shutting down OpenTelemetry tracer provider")
		if serr := m.traceProvider.Shutdown(ctx); serr != nil {
			err = errors.Join(err, fmt.Errorf("failed to shut down tracer provider: %w", serr))
		}
	}
	if m.traceExporter != nil {
		log.Debug("Shutting down OpenTelemetry trace exporter")
		if serr := m.traceExporter.Shutdown(ctx); serr != nil {
			err = errors.Join(err, fmt.Errorf("failed to shut down trace exporter: %w", serr))
		}
	}
	if m.logProvider != nil {
		log.Debug("Shutting down OpenTelemetry log provider")
		if serr := m.logProvider.Shutdown(ctx); serr != nil {
			err = errors.Join(err, fmt.Errorf("failed to shut down log provider: %w", serr))
		}
	}
	if m.logExporter != nil {
		log.Debug("Shutting down OpenTelemetry log exporter")
		if serr := m.logExporter.Shutdown(ctx); serr != nil {
			err = errors.Join(err, fmt.Errorf("failed to shut down log exporter: %w", serr))
		}
	}

	return err
}

// ShutdownCleanup performs final cleanup
func (m *Manager) ShutdownCleanup(ctx context.Context) error {
	return nil
}
