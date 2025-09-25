package grpc

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/pkg/servers"
)

func NewServer(router http.Handler, c APIServerConfig) servers.Server {
	return servers.NewHttpServer("api", router, servers.Config{
		Port: c.Port,
		Host: c.Host,
	})
}
