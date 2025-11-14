package components

import (
	"context"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/infrastructure/health"
	"github.com/DaanV2/mechanus/server/infrastructure/lifecycle"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/servers"
	"github.com/DaanV2/mechanus/server/infrastructure/tracing"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/grpc"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/http"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/mdns"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/websocket"
)

type Server struct {
	Manager    *servers.Manager
	Users      *application.UserService
	DB         *persistence.DB
	Components *lifecycle.Manager
}

func CreateWebServer(conf http.ServerConfig, healthChecker health.HealthCheck, readyChecker health.ReadyCheck, tracingConf tracing.Config) servers.Server {
	router := http.WebRouter(conf, healthChecker, readyChecker)

	return http.NewServer(conf.Config, router, tracingConf)
}

func CreateAPIServer(apiConfig grpc.APIServerConfig, websocketHandler *websocket.WebsocketHandler, rpcs grpc.RPCS, tracingConf tracing.Config) servers.Server {
	grpcrouter := grpc.NewRouter(rpcs, tracingConf)
	webrouter := websocket.NewWebsocketRouter(websocketHandler)
	router := http.NewWebsocketSplitter(webrouter, grpcrouter)

	return grpc.NewServer(router, apiConfig, tracingConf)
}

func CreateMDNSServer(ctx context.Context, conf mdns.ServerConfig) (*mdns.Server, error) {
	return mdns.NewServer(ctx, conf)
}
