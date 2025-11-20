package telemetry

import (
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/DaanV2/mechanus/server/mechanus"
	"github.com/charmbracelet/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// TraceHttpMiddleware wraps an HTTP handler with OpenTelemetry instrumentation
func TraceHttpMiddleware(cfg *Config, next http.Handler) http.Handler {
	if !cfg.Enabled {
		// Return the handler as-is if telemetry is disabled
		return next
	}

	// Wrap with otelhttp instrumentation
	return otelhttp.NewHandler(
		next,
		"server",
		otelhttp.WithServerName(mechanus.SERVICE_NAME),
	)
}

// TraceGRPCMiddleware returns a Connect interceptor that adds OpenTelemetry telemetry to gRPC requests
func TraceGRPCMiddleware(cfg *Config) connect.Interceptor {
	if !cfg.Enabled {
		return nil
	}

	otelInterceptor, err := otelconnect.NewInterceptor()
	if err != nil {
		log.Fatal("failed to create opentelemetry interceptor", "error", err)
	}

	return otelInterceptor
}
