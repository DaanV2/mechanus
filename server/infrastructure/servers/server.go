package servers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/DaanV2/mechanus/server/infrastructure/lifecycle"
	"github.com/charmbracelet/log"
)

var (
	_ lifecycle.ShutdownCleanup = &Server{}
)

type Config struct {
	Port uint16
	Host string
}

type Server struct {
	name   string
	server *http.Server
}

func NewServer(name string, router http.Handler, conf Config, opts ...Option) *Server {
	result := &Server{
		name: name,
		server: &http.Server{
			Addr:              fmt.Sprintf("%s:%v", conf.Host, conf.Port),
			Handler:           router,
			ReadHeaderTimeout: time.Second * 10,
		},
	}

	for _, opt := range opts {
		opt.apply(result.server)
	}

	return result
}

func (s *Server) Listen() {
	log.Infof("Starting %s server: http://%s", s.name, s.server.Addr)

	err := s.server.ListenAndServe()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}

		log.Errorf("error listening for server: %s %s => %s", s.name, s.server.Addr, err)
	}
}

func (s *Server) Shutdown(ctx context.Context) {
	err := s.server.Shutdown(ctx)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}

		log.Errorf("error listening for server: %s %s => %v", s.name, s.server.Addr, err)
	}
}

// ShutdownCleanup implements lifecycle.ShutdownCleanup.
func (s *Server) ShutdownCleanup(ctx context.Context) error {
	return s.server.Close()
}
