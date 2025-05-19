package components

import (
	"github.com/DaanV2/mechanus/server/internal/grpc"
	"github.com/DaanV2/mechanus/server/internal/web"
	"github.com/DaanV2/mechanus/server/pkg/servers"
)

func ServerManager(servs ...servers.Server) *servers.Manager {
	manager := &servers.Manager{}
	manager.Register(servs...)

	return manager
}

func WebServer(serv web.WEBServices) servers.Server {
	router := web.WebRouter(serv)

	return web.NewServer(router)
}

func APIServer(rpcs grpc.RPCS) servers.Server {
	router := grpc.NewRouter(rpcs)
	conf := grpc.GetConfig()

	return grpc.NewServer(router, conf)
}
