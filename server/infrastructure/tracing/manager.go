package tracing

import (
	"context"

	"github.com/charmbracelet/log"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Manager handles the lifecycle of the OpenTelemetry tracer provider
type Manager struct {
	provider *sdktrace.TracerProvider
}

// NewManager creates a new tracing manager
func NewManager(provider *sdktrace.TracerProvider) *Manager {
	return &Manager{
		provider: provider,
	}
}

// AfterInitialize is called during the initialization phase
func (m *Manager) AfterInitialize(ctx context.Context) error {
	if m.provider != nil {
		log.Debug("OpenTelemetry tracing initialized")
	}

	return nil
}

// BeforeShutdown is called before shutdown begins
func (m *Manager) BeforeShutdown(ctx context.Context) error {
	return nil
}

// AfterShutDown is called after shutdown completes
func (m *Manager) AfterShutDown(ctx context.Context) error {
	if m.provider != nil {
		log.Debug("Shutting down OpenTelemetry tracer provider")
		if err := Shutdown(ctx, m.provider); err != nil {
			log.Error("Failed to shutdown tracer provider", "error", err)

			return err
		}
	}

	return nil
}

// ShutdownCleanup performs final cleanup
func (m *Manager) ShutdownCleanup(ctx context.Context) error {
	return nil
}
