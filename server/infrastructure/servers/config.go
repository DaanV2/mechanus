package servers

import (
	"github.com/DaanV2/mechanus/server/infrastructure/config"
)

var (
	ServerConfigSet        = config.New("server")
	ServerHostFlag         = ServerConfigSet.String("server.host", "", "What host to bind on, such as: \"\", \"localhost\" or \"0.0.0.0\"")
	ServerPortFlag         = ServerConfigSet.UInt16("server.port", 8080, "The port to server web traffic to")
	ServerStaticFolderFlag = ServerConfigSet.String("server.static.folder", "/web", "The default files to serve")
	ServerTLSCertFlag      = ServerConfigSet.String("server.tls.cert", "", "Path to a PEM certificate (chain) file, when set together with server.tls.key the server serves HTTPS")
	ServerTLSKeyFlag       = ServerConfigSet.String("server.tls.key", "", "Path to the PEM private key file of server.tls.cert")
	ServerTrustProxyFlag   = ServerConfigSet.Bool("server.trust-proxy-headers", false, "Trust the X-Forwarded-Proto, X-Forwarded-Host and X-Forwarded-For headers of a reverse proxy in front of the server. Only enable when every request goes through that proxy")
)

type ServerConfig struct {
	Config
	StaticFolder string
	// TrustProxyHeaders applies the X-Forwarded-* headers of a reverse proxy to requests
	TrustProxyHeaders bool
}

func GetServerConfig() *ServerConfig {
	return &ServerConfig{
		Config: Config{
			Port:        ServerPortFlag.Value(),
			Host:        ServerHostFlag.Value(),
			TLSCertFile: ServerTLSCertFlag.Value(),
			TLSKeyFile:  ServerTLSKeyFlag.Value(),
		},
		StaticFolder:      ServerStaticFolderFlag.Value(),
		TrustProxyHeaders: ServerTrustProxyFlag.Value(),
	}
}
