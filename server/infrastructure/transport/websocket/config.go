package websocket

import (
	"time"

	"github.com/DaanV2/mechanus/server/infrastructure/config"
)

var (
	WebsocketConfigSet       = config.New("server.websocket")
	SessionCheckIntervalFlag = WebsocketConfigSet.Duration("server.websocket.session-check-interval", time.Minute, "How often the session of a logged in user on a websocket is checked for revocation, connections are always closed once their token expires")
)

type WebsocketConfig struct {
	// How often the JTI of a user connection is checked for revocation.
	SessionCheckInterval time.Duration
}

func GetWebsocketConfig() WebsocketConfig {
	return WebsocketConfig{
		SessionCheckInterval: SessionCheckIntervalFlag.Value(),
	}
}
