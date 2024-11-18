package config

import "github.com/spf13/pflag"

type PortIpConfig struct {
	Port Flag[int]
	Host Flag[string]
}

type GRPCConfig struct {
	PortIpConfig
	Reflection Flag[bool]
}

type WebConfig struct {
	PortIpConfig              // The config for the webserver
	Folder       Flag[string] // The folder to provide files out of
}

type APIServerConfig struct {
	GRPC GRPCConfig
	Rest PortIpConfig
	Web  WebConfig
}

var APIServer = &APIServerConfig{
	GRPC: GRPCConfig{
		PortIpConfig: PortIpConfig{
			Port: Int("api.grpc.port", 8090, "The port for the server"),
			Host: String("api.grpc.host", "0.0.0.0", "The host address to bind on"),
		},
		Reflection: Bool("api.grpc.reflection", true, "Whenever or not to turn on reflection for grpc"),
	},
	Rest: PortIpConfig{
		Port: Int("api.rest.port", 8091, "The port for the server"),
		Host: String("api.rest.host", "0.0.0.0", "The host address to bind on"),
	},
	Web: WebConfig{
		PortIpConfig: PortIpConfig{
			Port: Int("web.port", 8080, "The web port for the server, this provides all the files and entry points for users"),
			Host: String("web.host", "0.0.0.0", "The host address to bind on"),
		},
	},
}

func (c *APIServerConfig) AddToSet(set *pflag.FlagSet) {
	c.GRPC.Host.AddToSet(set)
	c.GRPC.Port.AddToSet(set)
	c.GRPC.Reflection.AddToSet(set)

	c.Rest.Host.AddToSet(set)
	c.Rest.Port.AddToSet(set)

	c.Web.Port.AddToSet(set)
	c.Web.Host.AddToSet(set)
	c.Web.Folder.AddToSet(set)
}
