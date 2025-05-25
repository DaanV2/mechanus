package mdns

import (
	"errors"
	"net"
	"runtime"

	"github.com/DaanV2/mechanus/server/mechanus/constants"
	"github.com/DaanV2/mechanus/server/pkg/config"
)

var (
	MDNSConfig      = config.New("mdns").WithValidate(validateServerConfig)
	HostNameFlag    = MDNSConfig.String("mdns.hostname", constants.SERVICE_NAME, "The host name to broadcast on")
	ServiceTypeFlag = MDNSConfig.String("mdns.servicetype", "_http._tcp.local", "The MDNS type to broadcast as")
	IPV6Flag        = MDNSConfig.Bool("mdns.ipv6", runtime.GOOS != "windows", "Whenever or not to support ipv6 as well")
)

type ServerConfig struct {
	IFace       *net.Interface
	HostName    string
	ServiceType string
	Port        uint16
	IPV6        bool
}

func validateServerConfig(c *config.Config) error {
	var err error

	if c.GetString("mdns.hostname") == "" {
		err = errors.Join(err, errors.New("host name for mdns needs to be something"))
	}
	if c.GetString("mdns.servicetype") == "" {
		err = errors.Join(err, errors.New("service type for mdns needs to be something"))
	}

	

	return err
}

func GetServerConfig(webport uint16) ServerConfig {
	return ServerConfig{
		IFace:       nil,
		HostName:    HostNameFlag.Value(),
		ServiceType: ServiceTypeFlag.Value(), // http service
		Port:        webport,
		IPV6:        IPV6Flag.Value(),
	}
}
