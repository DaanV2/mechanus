package grpc

import (
	"net/http"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/pkg/authenication"
	"github.com/DaanV2/mechanus/server/pkg/generics"
	"github.com/DaanV2/mechanus/server/pkg/grpc/gen/users/v1/usersv1connect"
	grpc_middleware "github.com/DaanV2/mechanus/server/pkg/grpc/middleware"
	"github.com/charmbracelet/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type GRPCServices struct {
	Login usersv1connect.LoginServiceHandler
	User  usersv1connect.UserServiceHandler
	JWT   *authenication.JWTService
}

func NewRouter(services GRPCServices) http.Handler {
	router := http.NewServeMux()

	opts := []connect.HandlerOption{
		connect.WithInterceptors(
			&MetadataInterceptor{},
			grpc_middleware.NewJWTMiddleware(services.JWT),
		),
	}

	RegisterService(router, usersv1connect.NewLoginServiceHandler, services.Login, opts...)
	RegisterService(router, usersv1connect.NewUserServiceHandler, services.User, opts...)

	return h2c.NewHandler(router, &http2.Server{})
}

func RegisterService[T any](router *http.ServeMux, create func(data T, opts ...connect.HandlerOption) (string, http.Handler), input T, opts ...connect.HandlerOption) {
	log.Debug("registering grpc service", "service", generics.NameOf[T]())
	path, handler := create(input, opts...)
	router.Handle(path, handler)
}
