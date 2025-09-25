package components

import (
	"context"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/grpc"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/http"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/mdns"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/websocket"
	"github.com/DaanV2/mechanus/server/pkg/servers"
)

type Server struct {
	Manager    *servers.Manager
	Users      *application.UserService
	DB         *persistence.DB
	Components *application.ComponentManager
}

func createServerManager(ctx context.Context, rpcs grpc.RPCS, websocketHandler *websocket.WebsocketHandler, serv http.WEBServices) (*servers.Manager, error) {
	wconf := http.GetConfig()
	apiconf := grpc.GetAPIServerConfig()
	mconf := mdns.GetServerConfig(wconf.Port)
	s, err := MDNSServer(ctx, mconf)
	if err != nil {
		return nil, err
	}

	return ServerManager(
		APIServer(apiconf, websocketHandler, rpcs),
		WebServer(wconf, serv),
		s,
	), nil
}

func ServerManager(servs ...servers.Server) *servers.Manager {
	manager := &servers.Manager{}
	manager.Register(servs...)

	return manager
}

func WebServer(conf http.ServerConfig, serv http.WEBServices) servers.Server {
	router := http.WebRouter(conf, serv)

	return http.NewServer(conf.Config, router)
}

func APIServer(apiConfig grpc.APIServerConfig, websocketHandler *websocket.WebsocketHandler, rpcs grpc.RPCS) servers.Server {
	grpcrouter := grpc.NewRouter(rpcs)
	webrouter := websocket.NewWebsocketRouter(websocketHandler)
	router := http.NewWebsocketSplitter(webrouter, grpcrouter)

	return grpc.NewServer(router, apiConfig)
}

func MDNSServer(ctx context.Context, conf mdns.ServerConfig) (*mdns.Server, error) {
	return mdns.NewServer(ctx, conf)
}
