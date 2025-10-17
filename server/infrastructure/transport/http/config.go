package http

import (
	"github.com/DaanV2/mechanus/server/infrastructure/config"
	"github.com/DaanV2/mechanus/server/infrastructure/servers"
)

var (
	WebConfig        = config.New("web")
	HostFlag         = WebConfig.String("web.host", "", "What host to bind on, such as: \"\", \"localhost\" or \"0.0.0.0\"")
	PortFlag         = WebConfig.UInt16("web.port", 8080, "The port to server web traffic to")
	StaticFolderFlag = WebConfig.String("web.static.folder", "/web", "The default files to serve")
)

type ServerConfig struct {
	servers.Config
	StaticFolder string
}

func GetWebConfig() ServerConfig {
	return ServerConfig{
		Config: servers.Config{
			Port: PortFlag.Value(),
			Host: HostFlag.Value(),
		},
		StaticFolder: StaticFolderFlag.Value(),
	}
}
