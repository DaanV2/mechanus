package grpc

import (
	"net/http"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/pkg/authenication"
	"github.com/DaanV2/mechanus/server/pkg/grpc/gen/screens/v1/screensv1connect"
	"github.com/DaanV2/mechanus/server/pkg/grpc/gen/users/v1/usersv1connect"
	grpc_handlers "github.com/DaanV2/mechanus/server/pkg/grpc/handlers"
	grpc_interceptors "github.com/DaanV2/mechanus/server/pkg/grpc/interceptors"
	"github.com/charmbracelet/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type RPCS struct {
	Login  usersv1connect.LoginServiceHandler
	User   usersv1connect.UserServiceHandler
	Screen screensv1connect.ScreensServiceHandler
	JWT    *authenication.JWTService
	CORS   *grpc_handlers.CORSHandler
}

func NewRouter(services RPCS) http.Handler {
	router := http.NewServeMux()

	opts := []connect.HandlerOption{
		connect.WithInterceptors(
			&grpc_interceptors.LoggingInterceptor{},
			grpc_interceptors.NewJWTMiddleware(services.JWT),
		),
	}

	RegisterService(router, usersv1connect.NewLoginServiceHandler, services.Login, opts...)
	RegisterService(router, usersv1connect.NewUserServiceHandler, services.User, opts...)
	RegisterService(router, screensv1connect.NewScreensServiceHandler, services.Screen, opts...)

	// Wrap the router with CORS middleware before h2c
	return h2c.NewHandler(grpc_handlers.Wrap(services.CORS, router), &http2.Server{})
}

func RegisterService[T any](router *http.ServeMux, create func(data T, opts ...connect.HandlerOption) (string, http.Handler), input T, opts ...connect.HandlerOption) {
	path, handler := create(input, opts...)
	log.Debug("registering grpc service", "service", path)

	router.Handle(path, handler)
}
