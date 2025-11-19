package websocket

import "github.com/DaanV2/mechanus/server/infrastructure/config"

var (
	WebsocketConfigSet = config.New("server.websocket")
)

type WebsocketConfig struct {
}

func GetWebsocketConfig() WebsocketConfig {
	return WebsocketConfig{}
}
