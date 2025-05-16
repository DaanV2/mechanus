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

type ServerConfig struct {
	Port int
	Host string
}

type Server struct {
	server *http.Server
}

func NewHttpServer(router http.Handler, conf ServerConfig) *Server {
	return &Server{
		server: &http.Server{
			Addr:              fmt.Sprintf("%s:%v", conf.Host, conf.Port),
			Handler:           middleware.Logging(router),
			ReadHeaderTimeout: time.Second * 10,
		},
	}
}

func (s *Server) Listen() {
	log.Infof("Starting http server: http://%s", s.server.Addr)

	err := s.server.ListenAndServe()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}

		log.Errorf("error listening for server: %s => %s", s.server.Addr, err)
	}
}

func (s *Server) Shutdown(ctx context.Context) {
	err := s.server.Shutdown(ctx)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}

		log.Errorf("error listening for server: %s => %v", s.server.Addr, err)
	}
}
