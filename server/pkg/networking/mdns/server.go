package mdns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"

	"github.com/DaanV2/mechanus/server/internal/logging"
	"github.com/DaanV2/mechanus/server/mechanus/constants"
)

type ServerConfig struct {
	IFace       *net.Interface
	HostName    string
	ServiceType string
	Port        int
	IPV6        bool
}

type Server struct {
	conf   *ServerConfig
	logger logging.Enriched

	conns map[string]*serverConn
}

func NewServerConfig(webport int) *ServerConfig {
	return &ServerConfig{
		IFace:       nil,
		HostName:    constants.SERVICE_NAME,
		ServiceType: "_http._tcp.local", // http service
		Port:        webport,
		IPV6:        runtime.GOOS != "windows",
	}
}

func NewServer(conf *ServerConfig) (*Server, error) {
	serv := &Server{
		conf:   conf,
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

	l := s.logger.With("network", network)
	l.From(context.Background()).Debug("started upd listener")

	s.conns[ip] = &serverConn{
		conf:   s.conf,
		logger: l,
		conn:   conn,
	}
	return nil
}

func (s *Server) Start() {
	for _, conn := range s.conns {
		go conn.Start()
	}
}
