package web

import (
	"net/http"

	http_middleware "github.com/DaanV2/mechanus/server/pkg/http/middleware"
	"github.com/DaanV2/mechanus/server/pkg/servers"
)

func NewServer(conf servers.Config, router http.Handler) servers.Server {
	return servers.NewHttpServer("web", http_middleware.Logging(router), conf)
}
