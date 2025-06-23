//go:build wireinject
// +build wireinject

package components

import (
	"context"

	"github.com/DaanV2/mechanus/server/internal/grpc"
	"github.com/DaanV2/mechanus/server/internal/web"
	"github.com/DaanV2/mechanus/server/pkg/application"
	"github.com/DaanV2/mechanus/server/pkg/authenication"
	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/DaanV2/mechanus/server/pkg/grpc/gen/users/v1/usersv1connect"
	grpc_handlers "github.com/DaanV2/mechanus/server/pkg/grpc/handlers"
	"github.com/DaanV2/mechanus/server/pkg/grpc/rpcs/rpcs_users"
	"github.com/DaanV2/mechanus/server/pkg/networking/mdns"
	"github.com/DaanV2/mechanus/server/pkg/servers"
	user_service "github.com/DaanV2/mechanus/server/pkg/services/users"
	"github.com/DaanV2/mechanus/server/pkg/storage"

	"github.com/google/wire"
)

type Server struct {
	Manager    *servers.Manager
	Users      *user_service.Service
	DB         *database.DB
	Components *application.ComponentManager
}

func BuildServer(ctx context.Context) (*Server, error) {
	wire.Build(
		dbSet,
		servicesSet,
		handlersSet,

		createServerManager,

		wire.Struct(new(grpc.RPCS), "*"),
		wire.Struct(new(web.WEBServices), "*"),
		wire.Struct(new(Server), "*"),
	)

	return &Server{}, nil
}

var dbSet = wire.NewSet(
	SetupDatabase,
	GetDatabaseOptions,
)

var handlersSet = wire.NewSet(
	grpc_handlers.GetCORSConfig,
	grpc_handlers.NewCORSHandler,
)

var servicesSet = wire.NewSet(
	application.NewComponentManager,

	rpcs_users.NewLoginService,
	rpcs_users.NewUserService,
	wire.Bind(new(usersv1connect.LoginServiceHandler), new(*rpcs_users.LoginService)),
	wire.Bind(new(usersv1connect.UserServiceHandler), new(*rpcs_users.UserService)),

	user_service.NewService,
	authenication.NewJWTService,
	authenication.NewJTIService,
	NewKeyManager,
	provideKeyStorage,
)

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

func provideKeyStorage(db *database.DB) storage.StorageProvider[*authenication.KeyData] {
	return storage.DBStorage[*authenication.KeyData](db)
}

