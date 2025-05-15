package grpc

import "github.com/DaanV2/mechanus/server/pkg/config"

var (
	APIConfig = config.New("api")
	HostFlag  = APIConfig.String("api.host", "", "What host to bind on, such as: \"\", \"localhost\" or \"0.0.0.0\"")
	PortFlag  = APIConfig.Int("api.port", 8666, "The port to server api traffic to")
)
