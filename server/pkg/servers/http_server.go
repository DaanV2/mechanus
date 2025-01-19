package servers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/DaanV2/mechanus/server/pkg/servers/middleware"
	"github.com/charmbracelet/log"
)

type HttpServerConfig struct {
	Port int
	Host string
}

type HttpServer struct {
	server *http.Server
}

func NewHttpServer(router http.Handler, conf HttpServerConfig) *HttpServer {
	return &HttpServer{
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%v", conf.Host, conf.Port),
			Handler: middleware.Logging(router),
		},
	}
}

func (s *HttpServer) Listen() {
	log.Infof("Starting http server: %s", s.server.Addr)

	err := s.server.ListenAndServe()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}

		log.Errorf("error listening for server: %s => %s", s.server.Addr, err)
	}
}

func (s *HttpServer) Shutdown() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second * 15))
	defer cancel()

	err := s.server.Shutdown(ctx)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}

		log.Errorf("error listening for server: %s => %v", s.server.Addr, err)
	}
}
