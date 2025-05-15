package servers

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	"github.com/charmbracelet/log"
)

type Manager struct {
	servers []*Server
	started bool
}

func (m *Manager) Register(server ...*Server) {
	m.servers = append(m.servers, server...)

	if m.started {
		log.Fatal("server manager has already started, but a server(s) is added")
	}
}

// Start is a non blocking operation, whereby all the api servers are started and processed
func (m *Manager) Start() {
	m.started = true
	log.Info("starting the servers")

	for _, server := range m.servers {
		go server.Listen()
	}
}

// Stop is a blocking operation, that gracefully shutdowns all servers, or stops them after 30 seconds
// If another kill, int or quit signal is caught, then its also cancelled
func (m *Manager) Stop(ctx context.Context) {
	m.started = false
	log.Info("shutting down servers...")
	defer log.Info("servers stopped")

	// Register a signal catcher
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	defer cancel()

	wg := &sync.WaitGroup{}

	for _, server := range m.servers {
		wg.Add(1)
		go m.stop(ctx, server, wg)
	}

	wg.Wait()
}

func (m *Manager) stop(ctx context.Context, server *Server, wg *sync.WaitGroup) {
	defer wg.Done()

	server.Shutdown(ctx)
}
