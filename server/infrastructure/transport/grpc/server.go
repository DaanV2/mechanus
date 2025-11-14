package grpc

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/infrastructure/servers"
	"github.com/DaanV2/mechanus/server/infrastructure/tracing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewServer(router http.Handler, c APIServerConfig, tracingConf tracing.Config) servers.Server {
	// Wrap with OpenTelemetry instrumentation if enabled
	handler := router
	if tracingConf.Enabled {
		handler = otelhttp.NewHandler(router, "grpc.server")
	}
	
	return servers.NewHttpServer("api", handler, servers.Config{
		Port: c.Port,
		Host: c.Host,
	})
}
