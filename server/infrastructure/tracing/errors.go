package tracing

import "github.com/charmbracelet/log"

type otelErrorHandler struct{}

func (o *otelErrorHandler) Handle(err error) {
	log.Error("error during otel", "error", err)
}
