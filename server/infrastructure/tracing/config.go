package tracing

import (
	"github.com/DaanV2/mechanus/server/infrastructure/config"
)

var (
	// TracingConfigSet contains all configuration flags for OpenTelemetry tracing
	TracingConfigSet = config.New("tracing")
	// EnabledFlag controls whether OpenTelemetry tracing is enabled
	EnabledFlag = TracingConfigSet.Bool("otel.enabled", false, "Enable OpenTelemetry tracing")
	// EndpointFlag specifies the OTLP collector endpoint
	EndpointFlag = TracingConfigSet.String("otel.endpoint", "localhost:4318", "OpenTelemetry collector endpoint (OTLP HTTP)")
	// ServiceNameFlag specifies the service name for traces
	ServiceNameFlag = TracingConfigSet.String("otel.service-name", "mechanus-server", "Service name for OpenTelemetry traces")
	// InsecureFlag controls whether to use insecure connection to OTLP collector
	InsecureFlag = TracingConfigSet.Bool("otel.insecure", true, "Use insecure connection to OTLP collector")
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
