package http

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/infrastructure/servers"
	"github.com/DaanV2/mechanus/server/infrastructure/tracing"
)

func NewServer(conf servers.Config, router http.Handler, tracingConf tracing.Config) servers.Server {
	// Apply middleware in order: OTel -> Logging
	handler := OtelMiddleware(tracingConf)(router)
	handler = Logging(handler)
	
	return servers.NewHttpServer("web", handler, conf)
}
