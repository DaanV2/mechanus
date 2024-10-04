package mdns

import (
	"errors"
	"net"

	"github.com/DaanV2/mechanus/server/pkg/config"
	"github.com/charmbracelet/log"
	"github.com/miekg/dns"
)

type discoverServiceOptions struct {
	ipv4 bool
	ipv6 bool
}

type DiscoverService struct {
	options    *discoverServiceOptions
	ipv4Server *net.UDPConn
	ipv6Server *net.UDPConn
	logger     *log.Logger
}

func NewDiscoverService() (*DiscoverService, error) {
	service := &DiscoverService{
		options: &discoverServiceOptions{
			ipv4: config.MDNS.IPV4.Value(),
			ipv6: config.MDNS.IPV6.Value(),
		},
		ipv4Server: nil,
		ipv6Server: nil,
		logger:     log.WithPrefix("mdns"),
	}
	var (
		iface *net.Interface
		err   error
	)

	// ipv4
	if service.options.ipv4 {
		service.ipv4Server, err = net.ListenMulticastUDP("udp4", iface, &net.UDPAddr{
			IP:   net.ParseIP(MDNS_IPV4),
			Port: MDNS_PORT,
		})
		if err != nil {
			return service, err
		}
	}

	//ipv6
	if service.options.ipv6 {
		service.ipv6Server, err = net.ListenMulticastUDP("udp6", iface, &net.UDPAddr{
			IP:   net.ParseIP(MDNS_IPV6),
			Port: MDNS_PORT,
		})
		if err != nil {
			return service, err
		}
	}

	go service.server(service.ipv4Server)
	go service.server(service.ipv6Server)

	return service, nil
}

func (ds *DiscoverService) Close() error {
	ds.logger.Debug("closing mdns server")

	var err error
	if ds.ipv4Server != nil {
		err = errors.Join(err, ds.ipv4Server.Close())
	}
	if ds.ipv6Server != nil {
		err = errors.Join(err, ds.ipv6Server.Close())
	}

	return err
}

func (ds *DiscoverService) server(conn *net.UDPConn) {
	if conn == nil {
		return
	}
	ds.logger.Debug("starting mdns server", "address", conn.LocalAddr())

	buf := make([]byte, 65536)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			ds.logger.Warn("error reading from mdns server", "error", err, "address", addr)
			continue
		}
		if n == 0 {
			continue
		}

		msg, err := parsePacket(buf[:n])
		if err != nil {
			ds.logger.Warn("error in packet format", "error", err)
			continue
		}

		err = ds.handleQuery(&msg)
		if err != nil {
			ds.logger.Warn("error in hanlding message", "error", err)
			continue
		}
	}
}

func parsePacket(packet []byte) (dns.Msg, error) {
	var msg dns.Msg
	err := msg.Unpack(packet)

	return msg, err
}

func (ds *DiscoverService) handleQuery(query *dns.Msg) error {
	ds.logger.Debug("got mdns message", "query", query)


	return nil
}
