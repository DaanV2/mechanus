package grpc

import "github.com/DaanV2/mechanus/server/infrastructure/config"

var (
	APIServerConfigSet = config.New("api")
	APIServerHostFlag  = APIServerConfigSet.String("api.host", "", "What host to bind on, such as: \"\", \"localhost\" or \"0.0.0.0\"")
	APIServerPortFlag  = APIServerConfigSet.UInt16("api.port", 8666, "The port to server api traffic to")
)

type APIServerConfig struct {
	Port uint16
	Host string
}

func GetAPIServerConfig() APIServerConfig {
	return APIServerConfig{
		Port: APIServerPortFlag.Value(),
		Host: APIServerHostFlag.Value(),
	}
}
