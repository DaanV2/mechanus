package components

import (
	"context"

	"github.com/DaanV2/mechanus/server/internal/grpc"
	"github.com/DaanV2/mechanus/server/internal/web"
	"github.com/DaanV2/mechanus/server/pkg/application"
	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/DaanV2/mechanus/server/pkg/networking/mdns"
	"github.com/DaanV2/mechanus/server/pkg/servers"
	user_service "github.com/DaanV2/mechanus/server/pkg/services/users"
)

type Server struct {
	Manager    *servers.Manager
	Users      *user_service.Service
	DB         *database.DB
	Components *application.ComponentManager
}

func createServerManager(ctx context.Context, rpcs grpc.RPCS, serv web.WEBServices) (*servers.Manager, error) {
	wconf := web.GetConfig()
	gconf := grpc.GetConfig()
	mconf := mdns.GetServerConfig(wconf.Port)
	s, err := MDNSServer(ctx, mconf)
	if err != nil {
		return nil, err
	}

	return ServerManager(
		APIServer(gconf, rpcs),
		WebServer(wconf, serv),
		s,
	), nil
}

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
