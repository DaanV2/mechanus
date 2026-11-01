package components

import (
	"context"
	gohttp "net/http"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/infrastructure/assets"
	"github.com/DaanV2/mechanus/server/infrastructure/health"
	"github.com/DaanV2/mechanus/server/infrastructure/lifecycle"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/servers"
	"github.com/DaanV2/mechanus/server/infrastructure/telemetry"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/cors"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/http"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/mdns"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/routers"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/websocket"
	"github.com/DaanV2/mechanus/server/mechanus"
	"github.com/DaanV2/mechanus/server/proto/campaigns/v1/campaignsv1connect"
	"github.com/DaanV2/mechanus/server/proto/scenes/v1/scenesv1connect"
	"github.com/DaanV2/mechanus/server/proto/users/v1/usersv1connect"
)

// ServerComponents holds the main server and its core components.
type ServerComponents struct {
	Server     *servers.Server
	Users      *application.UserService
	DB         *persistence.DB
	Components *lifecycle.Manager
}

// CreateMDNSServer creates a new mDNS server with the provided configuration.
func CreateMDNSServer(ctx context.Context, conf mdns.ServerConfig) (*mdns.Server, error) {
	return mdns.NewServer(ctx, conf)
}

// RouterSetup contains the setup dependencies for creating a router.
type RouterSetup struct {
	HealthChecker    health.HealthCheck
	Interceptors     []connect.Interceptor
	ReadyChecker     health.ReadyCheck
	WebsocketHandler *websocket.WebsocketHandler
	Assets           gohttp.Handler // Serves scene and shared assets under assets.RoutePrefix
}

// RouterRPCS contains the gRPC service handlers for the router.
type RouterRPCS struct {
	Campaign campaignsv1connect.CampaignServiceHandler
	Login    usersv1connect.LoginServiceHandler
	Scene    scenesv1connect.SceneServiceHandler
	User     usersv1connect.UserServiceHandler
	Invite   usersv1connect.InviteServiceHandler
}

// RouterConfig contains the configuration for creating a router.
type RouterConfig struct {
	CORS    *cors.CORSConfig
	Server  *servers.ServerConfig
	Tracing *telemetry.Config
}

// CreateRouter creates an HTTP handler with all routes, middleware, and services configured.
func CreateRouter(setup *RouterSetup, rpcs *RouterRPCS, cfgs RouterConfig) (gohttp.Handler, error) {
	healthServ := health.NewHealthService(setup.HealthChecker)
	readyServ := health.NewReadyService(setup.ReadyChecker)

	grpcOpts := []connect.HandlerOption{
		connect.WithInterceptors(setup.Interceptors...),
	}

	reflecter := grpcreflect.NewStaticReflector(
		campaignsv1connect.CampaignServiceName,
		usersv1connect.LoginServiceName,
		scenesv1connect.SceneServiceName,
		usersv1connect.UserServiceName,
		usersv1connect.InviteServiceName,
		grpchealth.HealthV1ServiceName,
	)

	wrouter := websocket.NewWebsocketRouter(setup.WebsocketHandler)
	router := routers.CreateRouter(
		// gRPC
		routers.WithGrpcRoute(campaignsv1connect.NewCampaignServiceHandler, rpcs.Campaign, grpcOpts),
		routers.WithGrpcRoute(usersv1connect.NewLoginServiceHandler, rpcs.Login, grpcOpts),
		routers.WithGrpcRoute(scenesv1connect.NewSceneServiceHandler, rpcs.Scene, grpcOpts),
		routers.WithGrpcRoute(usersv1connect.NewUserServiceHandler, rpcs.User, grpcOpts),
		routers.WithGrpcRoute(usersv1connect.NewInviteServiceHandler, rpcs.Invite, grpcOpts),
		// gRPC utils
		routers.WithGrpcRoute(grpcreflect.NewHandlerV1, reflecter, grpcOpts),
		routers.WithGrpcRoute(grpcreflect.NewHandlerV1Alpha, reflecter, grpcOpts),
		routers.WithGrpcRoute(grpchealth.NewHandler, grpchealth.Checker(healthServ), grpcOpts),

		// http
		routers.WithHandle("/", gohttp.FileServer(gohttp.Dir(cfgs.Server.StaticFolder))),
		routers.WithHandle(assets.RoutePrefix, setup.Assets),
		// http utils
		routers.WithHandle("/health", healthServ),
		routers.WithHandle("/healthz", healthServ),
		routers.WithHandle("/ready", readyServ),
		routers.WithHandle("/readyz", readyServ),
	)

	hrouter := gohttp.Handler(router)
	hrouter = cors.NewCORSHandler(cfgs.CORS, hrouter)
	hrouter = http.NewWebsocketSplitter(wrouter, hrouter)
	hrouter = telemetry.TraceHttpMiddleware(cfgs.Tracing, hrouter)
	hrouter = logging.HttpMiddleware(hrouter)
	if cfgs.Server.TrustProxyHeaders {
		hrouter = http.NewForwardedHeaders(hrouter)
	}

	return hrouter, nil
}

// CreateServer creates a new HTTP server with the provided handler and configuration.
func CreateServer(router gohttp.Handler, conf servers.Config) *servers.Server {
	p := new(gohttp.Protocols)
	p.SetHTTP1(true)
	p.SetHTTP2(true)
	p.SetUnencryptedHTTP2(true)

	return servers.NewServer(
		mechanus.SERVICE_NAME,
		router,
		conf,
		servers.WithProtocols(p),
	)
}
