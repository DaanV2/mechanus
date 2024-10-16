package grpc

import (
	"net/http"

	"github.com/charmbracelet/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)


type Server struct {}


func NewServer() {
	mux := http.NewServeMux()

	err := http.ListenAndServe(
		"localhost:8080",
		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(mux, &http2.Server{}),
	)
	if err != nil {
		log.Error("error while serving traffic", "error", err)
	}
}