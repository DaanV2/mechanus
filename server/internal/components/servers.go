package components

import (
	"context"

	"github.com/DaanV2/mechanus/server/internal/grpc"
	"github.com/DaanV2/mechanus/server/internal/web"
	"github.com/DaanV2/mechanus/server/pkg/networking/mdns"
	"github.com/DaanV2/mechanus/server/pkg/servers"
)

func ServerManager(servs ...servers.Server) *servers.Manager {
	manager := &servers.Manager{}
	manager.Register(servs...)

	return manager
}

func WebServer(conf web.ServerConfig, serv web.WEBServices) servers.Server {
	router := web.WebRouter(conf, serv)

	return web.NewServer(conf.Config, router)
}


func APIServer(conf grpc.Config, rpcs grpc.RPCS) servers.Server {
	router := grpc.NewRouter(rpcs)

	return grpc.NewServer(router, conf)
}

func MDNSServer(ctx context.Context, conf mdns.ServerConfig) (*mdns.Server, error) {
	return mdns.NewServer(ctx, conf)
}