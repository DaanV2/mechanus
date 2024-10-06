package mdns

import (
	"os"

	"github.com/DaanV2/mechanus/server/pkg/config"
	xnet "github.com/DaanV2/mechanus/server/pkg/net"
	"github.com/charmbracelet/log"
	hash_mdns "github.com/hashicorp/mdns"
)

type Service struct {
	server *hash_mdns.Server
}

func NewService() (*Service, error) {
	// Setup our service export
	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	info := []string{"My awesome service"}
	service, err := hash_mdns.NewMDNSService(host, "_foobar._tcp", "", "", 8000, nil, info)
	if err != nil {
		return nil, err
	}

	mdnsConfig := &hash_mdns.Config{
		Zone: service,
	}
	ifaceName := config.MDNS.IFace.Value()
	if ifaceName != "" {
		iface, ierr := xnet.FindIFace(ifaceName)
		if ierr != nil {
			return nil, ierr
		}
		mdnsConfig.Iface = &iface
	}

	log.Debug("starting new mdns service", "zone", mdnsConfig.Zone, "iface", mdnsConfig.Iface)
	server, err := hash_mdns.NewServer(mdnsConfig)
	if err != nil {
		return nil, err
	}

	return &Service{
		server,
	}, nil
}

func (s *Service) Close() error {
	return s.server.Shutdown()
}
