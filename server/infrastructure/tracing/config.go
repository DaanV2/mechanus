package tracing

import (
	"github.com/DaanV2/mechanus/server/infrastructure/config"
)

var (
	// OtelConfigSet contains all configuration flags for OpenTelemetry tracing
	OtelConfigSet = config.New("otel")
	// EnabledFlag controls whether OpenTelemetry tracing is enabled
	EnabledFlag = OtelConfigSet.Bool("otel.enabled", false, "Enable OpenTelemetry tracing")
	// EndpointFlag specifies the OTLP collector endpoint
	EndpointFlag = OtelConfigSet.String("otel.endpoint", "localhost:4318", "OpenTelemetry collector endpoint (OTLP HTTP)")
	// ServiceNameFlag specifies the service name for traces
	ServiceNameFlag = OtelConfigSet.String("otel.service-name", "mechanus-server", "Service name for OpenTelemetry traces")
	// InsecureFlag controls whether to use insecure connection to OTLP collector
	InsecureFlag = OtelConfigSet.Bool("otel.insecure", true, "Use insecure connection to OTLP collector")
)

// GetConfig returns the current tracing configuration
func GetConfig() *Config {
	return &Config{
		Enabled:     EnabledFlag.Value(),
		Endpoint:    EndpointFlag.Value(),
		ServiceName: ServiceNameFlag.Value(),
		Insecure:    InsecureFlag.Value(),
	}
}

// Config holds the OpenTelemetry configuration
type Config struct {
	Enabled     bool
	Endpoint    string
	ServiceName string
	Insecure    bool
}
