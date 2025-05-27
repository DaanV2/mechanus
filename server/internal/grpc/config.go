package grpc

import "github.com/DaanV2/mechanus/server/pkg/config"

var (
	APIConfig = config.New("api")
	HostFlag  = APIConfig.String("api.host", "", "What host to bind on, such as: \"\", \"localhost\" or \"0.0.0.0\"")
	PortFlag  = APIConfig.UInt16("api.port", 8666, "The port to server api traffic to")
)

type Config struct {
	Port uint16
	Host string
}

func GetConfig() Config {
	return Config{
		Port: PortFlag.Value(),
		Host: HostFlag.Value(),
	}
}