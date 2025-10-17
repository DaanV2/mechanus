package grpc

import (
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/tracing"
	"github.com/DaanV2/mechanus/server/proto/users/v1/usersv1connect"
	"github.com/charmbracelet/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type RPCS struct {
	Login usersv1connect.LoginServiceHandler
	User  usersv1connect.UserServiceHandler
	JWT   *authentication.JWTService
	CORS  *CORSHandler
}

func NewRouter(services RPCS, tracingConf tracing.Config) http.Handler {
	router := http.NewServeMux()

	opts := []connect.HandlerOption{
		connect.WithInterceptors(
			&LoggingInterceptor{},
			NewAuthenticator(services.JWT),
		),
	}

	// Add OpenTelemetry interceptor if tracing is enabled
	if tracingConf.Enabled {
		otelInterceptor, err := otelconnect.NewInterceptor()
		if err == nil {
			opts = append(opts, connect.WithInterceptors(otelInterceptor))
		} else {
			log.Warn("Failed to create OpenTelemetry interceptor", "error", err)
		}
	}

	RegisterService(router, usersv1connect.NewLoginServiceHandler, services.Login, opts...)
	RegisterService(router, usersv1connect.NewUserServiceHandler, services.User, opts...)

	// Wrap the router with CORS middleware before h2c
	return h2c.NewHandler(CORSWarp(services.CORS, router), &http2.Server{})
}

func RegisterService[T any](router *http.ServeMux, create func(data T, opts ...connect.HandlerOption) (string, http.Handler), input T, opts ...connect.HandlerOption) {
	path, handler := create(input, opts...)
	log.Debug("registering grpc service", "service", path)

	router.Handle(path, handler)
}
