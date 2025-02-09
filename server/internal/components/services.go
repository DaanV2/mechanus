package components

import (
	grpc_users "github.com/DaanV2/mechanus/server/internal/grpc/users"
	"github.com/DaanV2/mechanus/server/pkg/authenication"
	"github.com/DaanV2/mechanus/server/pkg/grpc/gen/users/v1/usersv1connect"
	user_service "github.com/DaanV2/mechanus/server/pkg/services/users"
	"github.com/google/wire"
)

var servicesSet = wire.NewSet( // nolint:unused
	user_service.NewService,
	authenication.NewJTIService,
	authenication.NewJWTService,
	authenication.NewKeyManager,

	// grpc rpcs
	grpc_users.NewLoginService,
	grpc_users.NewUserService,
	wire.Bind( new(usersv1connect.LoginServiceHandler), new(*grpc_users.LoginService)),
	wire.Bind( new(usersv1connect.UserServiceHandler), new(*grpc_users.UserService)),
)
