package mdns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/DaanV2/mechanus/server/internal/logging"
	"github.com/DaanV2/mechanus/server/mechanus/constants"
	"github.com/DaanV2/mechanus/server/pkg/networking/mdns"
	"golang.org/x/net/dns/dnsmessage"
)

const (
	QUERY_OPCODE = dnsmessage.OpCode(0)
)

type serverConn struct {
	conf     *ServerConfig
	logger   logging.Enriched
	conn     *net.UDPConn
	ctx      context.Context
	hostname dnsmessage.Name
}

func (s *serverConn) Start() {
	buffer := make([]byte, 2048)
	s.ctx = context.Background()
	s.hostname = dnsmessage.MustNewName(s.conf.HostName + ".")
	logger := s.logger.From(s.ctx)

	go s.annouceLoop()

	for {
		n, src, err := s.conn.ReadFromUDP(buffer)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				logger.Debug("connection closed on upd server")
				return
			}

			logger.Error("error reading from UPD stream", "error", err, "source", src)
			continue
		}

		s.handlePacket(buffer[:n], src)
	}
}

func (s *serverConn) handlePacket(data []byte, src *net.UDPAddr) {
	logger := s.logger.With("source", src).From(s.ctx)

	var msg dnsmessage.Message
	err := msg.Unpack(data)
	if err != nil {
		logger.Error("couldn't read message", "error", err)
		return
	}

	if msg.Header.Response {
		return // ignore
	}

	// Ensure trailing dot
	var answers []dnsmessage.Resource

	for _, q := range msg.Questions {

		ans, cerr := s.checkQuestion(q)
		if cerr != nil {
			logger.Error("couldn't make answer to error", "error", cerr, "question", q)
		} else if len(ans) > 0 {
			logger.Debug("question", "name", q.Name.String(), "class", q.Class.String(), "type", q.Type.String())
			logger.Debug("answers", "answers", answers)
			answers = append(answers, ans...)
		}
	}

	if len(answers) <= 0 {
		return // nothing to answer
	}

	resp := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID:            msg.ID,
			Response:      true,
			Authoritative: true,
		},
		Questions: msg.Questions,
		Answers:   answers,
	}
	err = s.sendMsg(src, resp)
	if err != nil {
		logger.Error("failed to send mDNS response", "error", err)
	}
}

func (s *serverConn) checkQuestion(q dnsmessage.Question) ([]dnsmessage.Resource, error) {
	var answers []dnsmessage.Resource

	serviceType := dnsmessage.MustNewName(s.conf.ServiceType + ".")
	instanceName := dnsmessage.MustNewName(s.conf.HostName + "." + s.conf.ServiceType + ".")
	hostname := s.hostname

	switch {
	case q.Name == serviceType && q.Type == dnsmessage.TypePTR:
		answers = append(answers, mdns.BuildPTRRecord(serviceType, instanceName))

	case q.Name == instanceName && q.Type == dnsmessage.TypeSRV:
		answers = append(answers, mdns.BuildSRVRecord(instanceName, hostname, uint16(s.conf.Port)))

	case q.Name == instanceName && q.Type == dnsmessage.TypeTXT:
		answers = append(answers, mdns.BuildTXTRecord(instanceName, []string{"path=/", "version=1.0"}))
		
	case q.Name == hostname && q.Type == dnsmessage.TypeA:
		ip := s.getLocalIPv4()
		var ipd [4]byte
		copy(ipd[:], ip)
		answers = append(answers, mdns.BuildARecord(hostname, ipd))
	case strings.Contains(q.Name.String(), constants.SERVICE_NAME):
		s.logger.From(s.ctx).Debug("unknown question", "question", q)
	}

	return answers, nil
}

// Helper to get the first non-loopback IPv4 address
func (s *serverConn) getLocalIPv4() net.IP {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip != nil {
				return ip
			}
		}
	}
	return nil
}

func (s *serverConn) sendMsg(receiver *net.UDPAddr, msg dnsmessage.Message) error {
	s.logger.From(s.ctx).Debug("sending a dns message", "msgId", msg.Header.ID)
	packed, err := msg.Pack()
	if err != nil {
		return fmt.Errorf("error packing dns message: %w", err)
	}

	// Send the message
	_, err = s.conn.WriteToUDP(packed, receiver)
	if err != nil {
		return fmt.Errorf("error writing dns message: %w", err)
	}

	return nil
}

func (s *serverConn) annouceLoop() {
	// Announce every 30 seconds (or on startup/change)
	for {
		s.sendAnnouncement()

		select {
		case <-time.After(30 * time.Second):
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *serverConn) sendAnnouncement() {
	s.logger.From(s.ctx).Debug("sending a mdns announcement")
	serviceType := dnsmessage.MustNewName(s.conf.ServiceType + ".")
	instanceName := dnsmessage.MustNewName(s.conf.HostName + "." + s.conf.ServiceType + ".")
	hostname := s.hostname

	ip := s.getLocalIPv4()
	var ipd [4]byte
	copy(ipd[:], ip)

	records := []dnsmessage.Resource{
		mdns.BuildPTRRecord(serviceType, instanceName),
		mdns.BuildSRVRecord(instanceName, hostname, uint16(s.conf.Port)),
		mdns.BuildTXTRecord(instanceName, []string{"path=/", "version=1.0"}),
		mdns.BuildARecord(hostname, ipd),
	}

	msg := dnsmessage.Message{
		Header: dnsmessage.Header{
			Response:      true,
			Authoritative: true,
		},
		Answers: records,
	}

	// Multicast address for mDNS
	addr := &net.UDPAddr{
		IP:   net.ParseIP(MDNS_ADDRESS_IPV4),
		Port: MDNS_PORT,
	}
	err := s.sendMsg(addr, msg)
	if err != nil {
		s.logger.From(s.ctx).Error("couldn't send mdns annoucment", "error", err)
	}
}
