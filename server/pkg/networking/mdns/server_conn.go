package mdns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/DaanV2/mechanus/server/internal/logging"
	"github.com/DaanV2/mechanus/server/mechanus/constants"
	"golang.org/x/net/dns/dnsmessage"
)

const (
	QUERY_OPCODE = dnsmessage.OpCode(0)
)

type serverConn struct {
	ctx      context.Context
	conf     *ServerConfig
	logger   logging.Enriched
	conn     *net.UDPConn
	hostname dnsmessage.Name
}

func newServerCon(ctx context.Context, conf *ServerConfig, logger logging.Enriched, conn *net.UDPConn) *serverConn {
	return &serverConn{
		ctx:      ctx,
		conf:     conf,
		logger:   logger,
		conn:     conn,
		hostname: dnsmessage.MustNewName(conf.HostName + "."),
	}
}

func (s *serverConn) Listen() {
	buffer := make([]byte, 2048)
	logger := s.logger.From(s.ctx)

	n := s.conn.LocalAddr().Network()
	if n == "upd" || n == "upd4" {
		go s.announceLoop()
	}

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

func (s *serverConn) Shutdown(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	logger := s.logger.From(ctx)
	logger.Debug("shutting down mdns server...")

	err := s.conn.Close()
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			return // Stopped.
		}

		logger.Error("error shutting down mdns server", "error", err)
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

	if msg.Response {
		return // ignore
	}

	// Ensure trailing dot
	var answers []dnsmessage.Resource

	for _, q := range msg.Questions { //nolint:gocritic // mdns isn't performed that much
		ans, cerr := s.checkQuestion(&q)
		if cerr != nil {
			logger.Error("couldn't make answer to error", "error", cerr, "question", q)
		} else if len(ans) > 0 {
			logger.Debug("question", "name", q.Name.String(), "class", q.Class.String(), "type", q.Type.String())
			logger.Debug("answers", "answers", answers)
			answers = append(answers, ans...)
		}
	}

	if len(answers) == 0 {
		return // nothing to answer
	}

	// Send answers
	resp := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID:            msg.ID,
			Response:      true,
			Authoritative: true,
		},
		Questions: msg.Questions,
		Answers:   answers,
	}
	err = s.sendMsg(src, &resp)
	if err != nil {
		logger.Error("failed to send mDNS response", "error", err)
	}
}

func (s *serverConn) checkQuestion(q *dnsmessage.Question) ([]dnsmessage.Resource, error) {
	var answers []dnsmessage.Resource

	serviceType := dnsmessage.MustNewName(s.conf.ServiceType + ".")
	instanceName := dnsmessage.MustNewName(s.conf.HostName + "." + s.conf.ServiceType + ".")
	hostname := s.hostname

	switch {
	case q.Name == serviceType && q.Type == dnsmessage.TypePTR:
		answers = append(answers, BuildPTRRecord(serviceType, instanceName))

	case q.Name == instanceName && q.Type == dnsmessage.TypeSRV:
		answers = append(answers, BuildSRVRecord(instanceName, hostname, s.conf.Port))

	case q.Name == instanceName && q.Type == dnsmessage.TypeTXT:
		answers = append(answers, BuildTXTRecord(instanceName, []string{"path=/", "version=1.0"}))

	case q.Name == hostname && q.Type == dnsmessage.TypeA:
		ip := getIPBytes()
		answers = append(answers, BuildARecord(hostname, ip))

	case strings.Contains(q.Name.String(), constants.SERVICE_NAME):
		s.logger.From(s.ctx).Debug("unknown question", "question", q)
	}

	return answers, nil
}

func (s *serverConn) sendMsg(receiver *net.UDPAddr, msg *dnsmessage.Message) error {
	s.logger.From(s.ctx).Debug("sending a dns message", "msgId", msg.ID)
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

// announceLoop continulisious until the CTX is done
func (s *serverConn) announceLoop() {
	s.logger.From(s.ctx).Debug("starting mdns announcement loop")

	// Announce every 30 seconds (or on startup/change)
	for {
		err := s.sendAnnouncement()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			} else {
				s.logger.From(s.ctx).Error("couldn't send mdns annoucment", "error", err)
			}
		}

		select {
		case <-time.After(30 * time.Second):
		case <-s.ctx.Done():
			return
		}
	}
}

// sendAnnouncement make a UPD broadcast to all who want to listen about this service.
// This should be able
func (s *serverConn) sendAnnouncement() error {
	s.logger.From(s.ctx).Debug("sending a mdns announcement")
	serviceType := dnsmessage.MustNewName(s.conf.ServiceType + ".")
	instanceName := dnsmessage.MustNewName(s.conf.HostName + "." + s.conf.ServiceType + ".")
	hostname := s.hostname

	ip := getIPBytes()

	records := []dnsmessage.Resource{
		BuildPTRRecord(serviceType, instanceName),
		BuildSRVRecord(instanceName, hostname, s.conf.Port),
		BuildTXTRecord(instanceName, []string{"path=/", "version=1.0"}),
		BuildARecord(hostname, ip),
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

	return s.sendMsg(addr, &msg)
}

func getIPBytes() [4]byte {
	ip := getLocalIPv4()
	var ipd [4]byte
	copy(ipd[:], ip)

	return ipd
}

// getLocalIPv4 is a helper to get the first non-loopback IPv4 address
func getLocalIPv4() net.IP {
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
