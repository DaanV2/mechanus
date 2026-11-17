package telemetry

import "github.com/DaanV2/mechanus/server/infrastructure/logging"

type otelErrorHandler struct{}

func (o *otelErrorHandler) Handle(err error) {
	// Console only: an export error sent back through the exporter would just fail again
	logging.Console().Error("error during otel", "error", err)
}
