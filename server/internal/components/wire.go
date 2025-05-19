//go:build wireinject
// +build wireinject

package components

import (
	"github.com/DaanV2/mechanus/server/internal/grpc"
	"github.com/DaanV2/mechanus/server/internal/web"
	"github.com/DaanV2/mechanus/server/pkg/application"
	"github.com/DaanV2/mechanus/server/pkg/authenication"
	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/DaanV2/mechanus/server/pkg/grpc/gen/users/v1/usersv1connect"
	"github.com/DaanV2/mechanus/server/pkg/grpc/rpcs/rpcs_users"
	"github.com/DaanV2/mechanus/server/pkg/servers"
	user_service "github.com/DaanV2/mechanus/server/pkg/services/users"
	"github.com/DaanV2/mechanus/server/pkg/storage"

	"github.com/google/wire"
)

type Server struct {
	Manager *servers.Manager
	Users   *user_service.Service
	DB      *database.DB
}

func BuildServer() (*Server, error) {
	wire.Build(
		dbSet,
		servicesSet,

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

var servicesSet = wire.NewSet(
	application.NewComponentManager,

	rpcs_users.NewLoginService,
	rpcs_users.NewUserService,
	wire.Bind(new(usersv1connect.LoginServiceHandler), new(*rpcs_users.LoginService)),
	wire.Bind(new(usersv1connect.UserServiceHandler), new(*rpcs_users.UserService)),

	user_service.NewService,
	authenication.NewJWTService,
	authenication.NewJTIService,
	authenication.NewKeyManager,
	provideKeyStorage,
)

func createServerManager(rpcs grpc.RPCS, serv web.WEBServices) *servers.Manager {
	return ServerManager(
		APIServer(rpcs),
		WebServer(serv),
	)
}

func provideKeyStorage(db *database.DB) storage.StorageProvider[*authenication.KeyData] {
	return storage.DBStorage[*authenication.KeyData](db)
}
