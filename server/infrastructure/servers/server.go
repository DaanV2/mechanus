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
	// TLSCertFile and TLSKeyFile are the PEM files to serve HTTPS with, when either is empty the server serves plain HTTP.
	TLSCertFile string
	TLSKeyFile  string
}

// TLSEnabled returns whenever or not the server is configured to serve HTTPS.
func (c Config) TLSEnabled() bool {
	return c.TLSCertFile != "" && c.TLSKeyFile != ""
}

type Server struct {
	name   string
	conf   Config
	server *http.Server
}

func NewServer(name string, router http.Handler, conf Config, opts ...Option) *Server {
	result := &Server{
		name: name,
		conf: conf,
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
	var err error
	if s.conf.TLSEnabled() {
		log.Infof("Starting %s server: https://%s", s.name, s.server.Addr)
		err = s.server.ListenAndServeTLS(s.conf.TLSCertFile, s.conf.TLSKeyFile)
	} else {
		log.Infof("Starting %s server: http://%s", s.name, s.server.Addr)
		err = s.server.ListenAndServe()
	}
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
