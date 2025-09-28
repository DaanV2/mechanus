package http

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/infrastructure/servers"
)

func NewServer(conf servers.Config, router http.Handler) servers.Server {
	return servers.NewHttpServer("web", Logging(router), conf)
}
