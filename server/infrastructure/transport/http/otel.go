package http

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/infrastructure/tracing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// OtelMiddleware wraps an HTTP handler with OpenTelemetry instrumentation
func OtelMiddleware(cfg tracing.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if !cfg.Enabled {
			// Return the handler as-is if tracing is disabled
			return next
		}
		
		// Wrap with otelhttp instrumentation
		return otelhttp.NewHandler(next, "http.server")
	}
}
