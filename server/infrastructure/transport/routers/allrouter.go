package routers

import (
	"net/http"
)

// CreateRouter creates a new HTTP router with the given options.
func CreateRouter(opts ...Option) *http.ServeMux {
	router := http.NewServeMux()
	
	for _, opt := range opts {
		opt(router)
	}

	return router
}