package grpc

import (
	"net/http"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/pkg/grpc/gen/users/v1/usersv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func NewRouter(loginSvc usersv1connect.LoginServiceHandler) http.Handler {
	router := http.NewServeMux()

	RegisterService(router,  usersv1connect.NewLoginServiceHandler, loginSvc)

	return h2c.NewHandler(router, &http2.Server{})
}

func RegisterService[T any](router *http.ServeMux, create func(data T, opts ...connect.HandlerOption) (string, http.Handler), input T, opts ...connect.HandlerOption) {
	path, handler := create(input, opts...)
	router.Handle(path, handler)
}
