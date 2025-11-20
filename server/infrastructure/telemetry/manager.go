package telemetry

import (
	"context"
	"errors"
	"fmt"

	"github.com/DaanV2/mechanus/server/infrastructure/lifecycle"
	"github.com/charmbracelet/log"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var (
	_ lifecycle.AfterInitialize = &Manager{}
	_ lifecycle.AfterShutDown = &Manager{}
	_ lifecycle.BeforeShutdown = &Manager{}
	_ lifecycle.ShutdownCleanup = &Manager{}
)

// Manager handles the lifecycle of the OpenTelemetry tracer provider
type Manager struct {
	provider *sdktrace.TracerProvider
	exporter *otlptrace.Exporter
}

// NewManager creates a new telemetry manager
func NewManager() *Manager {
	return &Manager{
		provider: nil,
		exporter: nil,
	}
}

func (m *Manager) SetTraceProvider(provider *sdktrace.TracerProvider) { m.provider = provider }
func (m *Manager) SetExporter(exporter *otlptrace.Exporter) { m.exporter = exporter }

// AfterInitialize is called during the initialization phase
func (m *Manager) AfterInitialize(ctx context.Context) error {
	if m.provider != nil {
		log.Debug("OpenTelemetry telemetry initialized")
	}
	if m.exporter != nil {
		log.Debug("OpenTelemetry exporter initialized")
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

	if m.provider != nil {
		log.Debug("Shutting down OpenTelemetry tracer provider")
		if serr := m.provider.Shutdown(ctx); serr != nil {
			err = errors.Join(err, fmt.Errorf("failed to shut down tracer provider: %w", serr))
		}
	}
	if m.exporter != nil {
		log.Debug("Shutting down OpenTelemetry exporter")
		if serr := m.exporter.Shutdown(ctx); serr != nil {
			err = errors.Join(err, fmt.Errorf("failed to shut down exporter: %w", serr))
		}
	}

	return err
}

// ShutdownCleanup performs final cleanup
func (m *Manager) ShutdownCleanup(ctx context.Context) error {
	return nil
}
