package tracing

import (
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
)

// SetupLoggerBridge connects the charm logger with the OTEL log exporter
// This should be called after SetupLogging has been called
func SetupLoggerBridge() {
	logging.SetupOtelBridge()
}
