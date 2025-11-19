package tracing

import (
	"context"

	"github.com/charmbracelet/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Manager handles the lifecycle of the OpenTelemetry tracer and logger providers
type Manager struct {
	provider    *sdktrace.TracerProvider
	logProvider *sdklog.LoggerProvider
}

// NewManager creates a new tracing manager
func NewManager(provider *sdktrace.TracerProvider, logProvider *sdklog.LoggerProvider) *Manager {
	return &Manager{
		provider:    provider,
		logProvider: logProvider,
	}
}

// AfterInitialize is called during the initialization phase
func (m *Manager) AfterInitialize(ctx context.Context) error {
	if m.provider != nil {
		log.Debug("OpenTelemetry tracing initialized")
	}

	if m.logProvider != nil {
		log.Debug("OpenTelemetry logging initialized")
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
		if shutdownErr := Shutdown(ctx, m.provider); shutdownErr != nil {
			log.Error("Failed to shutdown tracer provider", "error", shutdownErr)
			err = shutdownErr
		}
	}

	if m.logProvider != nil {
		log.Debug("Shutting down OpenTelemetry logger provider")
		if shutdownErr := ShutdownLogging(ctx, m.logProvider); shutdownErr != nil {
			log.Error("Failed to shutdown logger provider", "error", shutdownErr)
			if err == nil {
				err = shutdownErr
			}
		}
	}

	return err
}

// ShutdownCleanup performs final cleanup
func (m *Manager) ShutdownCleanup(ctx context.Context) error {
	return nil
}
