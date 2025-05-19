package web

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/pkg/servers"
)

func NewServer(router http.Handler) servers.Server {
	return servers.NewHttpServer("web", router, servers.Config{
		Port: PortFlag.Value(),
		Host: HostFlag.Value(),
	})
}
