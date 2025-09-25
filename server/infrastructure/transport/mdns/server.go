package mdns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
)

// Server handles service discovery via mdns, it starts the
type Server struct {
	conf   *ServerConfig
	logger logging.Enriched
	ctx    context.Context

	conns map[string]*serverConn
}

func NewServer(ctx context.Context, conf ServerConfig) (*Server, error) {
	serv := &Server{
		ctx:    ctx,
		conf:   &conf,
		logger: logging.Enriched{}.WithPrefix("mdns"),
		conns:  make(map[string]*serverConn),
	}

	err := serv.registerServer("udp4", MDNS_ADDRESS_IPV4, MDNS_PORT)
	if conf.IPV6 {
		err = errors.Join(err, serv.registerServer("udp6", MDNS_ADDRESS_IPV6, MDNS_PORT))
	}

	return serv, err
}

func (s *Server) registerServer(network, ip string, port int) error {
	addr := &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}

	conn, err := net.ListenMulticastUDP(network, s.conf.IFace, addr)
	if err != nil {
		return fmt.Errorf("error setting up upd network on %s:%d => %w", ip, port, err)
	}

	s.conns[ip] = newServerCon(s.ctx, s.conf, s.logger.With("network", network), conn)

	return nil
}

// Listen implements [servers.Server.Listen]
func (s *Server) Listen() {
	s.logger.From(s.ctx).Info("starting mdns server")

	for _, conn := range s.conns {
		go conn.Listen()
	}
}

// Listen implements [servers.Server.Shutdown]
func (s *Server) Shutdown(ctx context.Context) {
	wg := &sync.WaitGroup{}

	for _, conn := range s.conns {
		wg.Add(1)

		go conn.Shutdown(ctx, wg)
	}

	wg.Wait()
}
