package servers

import "net/http"

type Option interface {
	apply(*http.Server)
}

type optionFn func(*http.Server)

func (f optionFn) apply(s *http.Server) {
	f(s)
}

// WithProtocol sets the allowed protocols for the HTTP server.
func WithProtocols(protos *http.Protocols) Option {
	return optionFn(func(s *http.Server) {
		s.Protocols = protos
	})
}