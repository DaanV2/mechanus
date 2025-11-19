package routers

import (
	"net/http"

	"connectrpc.com/connect"
	"github.com/charmbracelet/log"
)

type Option func(*http.ServeMux)

func WithGrpcRoute[T any](create func(data T, opts ...connect.HandlerOption) (string, http.Handler), input T, opts []connect.HandlerOption) Option {
	return func(router *http.ServeMux) {
		path, handler := create(input, opts...)
		log.Debug("registering grpc service: " + path)

		router.Handle(path, handler)
	}
}

func WithHandle(pattern string, handler http.Handler) Option {
	return func(router *http.ServeMux) {
		log.Debug("registering route: " + pattern)

		router.Handle(pattern, handler)
	}
}
