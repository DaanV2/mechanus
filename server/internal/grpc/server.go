package grpc

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/pkg/servers"
)

func NewServer(router http.Handler) *servers.Server {
	return servers.NewHttpServer(router, servers.ServerConfig{
		Port: PortFlag.Value(),
		Host: HostFlag.Value(),
	})
}
