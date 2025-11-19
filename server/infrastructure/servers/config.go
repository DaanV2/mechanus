package servers

import (
	"github.com/DaanV2/mechanus/server/infrastructure/config"
)

var (
	ServerConfigSet        = config.New("server")
	ServerHostFlag         = ServerConfigSet.String("server.host", "", "What host to bind on, such as: \"\", \"localhost\" or \"0.0.0.0\"")
	ServerPortFlag         = ServerConfigSet.UInt16("server.port", 8080, "The port to server web traffic to")
	ServerStaticFolderFlag = ServerConfigSet.String("server.static.folder", "/web", "The default files to serve")
)

type ServerConfig struct {
	Config
	StaticFolder string
}

func GetServerConfig() *ServerConfig {
	return &ServerConfig{
		Config: Config{
			Port: ServerPortFlag.Value(),
			Host: ServerHostFlag.Value(),
		},
		StaticFolder: ServerStaticFolderFlag.Value(),
	}
}