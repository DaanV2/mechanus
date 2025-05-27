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

type Config struct {
	Port uint16
	Host string
}

type HTTPServer struct {
	name   string
	server *http.Server
}

func NewHttpServer(name string, router http.Handler, conf Config) Server {
	return &HTTPServer{
		name: name,
		server: &http.Server{
			Addr:              fmt.Sprintf("%s:%v", conf.Host, conf.Port),
			Handler:           middleware.Logging(router),
			ReadHeaderTimeout: time.Second * 10,
		},
	}
}

func (s *HTTPServer) Listen() {
	log.Infof("Starting %s server: http://%s", s.name, s.server.Addr)

	err := s.server.ListenAndServe()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}

		log.Errorf("error listening for server: %s %s => %s", s.name, s.server.Addr, err)
	}
}

func (s *HTTPServer) Shutdown(ctx context.Context) {
	err := s.server.Shutdown(ctx)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}

		log.Errorf("error listening for server: %s %s => %v", s.name, s.server.Addr, err)
	}
}
