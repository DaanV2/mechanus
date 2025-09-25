package websocket

import "github.com/DaanV2/mechanus/server/infrastructure/config"

var (
	WebsocketConfigSet = config.New("api.websocket")
	// HostFlag        = WebsocketConfigSet.String("api.host", "", "What host to bind on, such as: \"\", \"localhost\" or \"0.0.0.0\"")
	// PortFlag        = WebsocketConfigSet.UInt16("api.port", 8666, "The port to server api traffic to")
)

type WebsocketConfig struct {
}

func GetWebsocketConfig() WebsocketConfig {
	return WebsocketConfig{}
}
