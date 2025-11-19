package logging

import (
	"net/http"

	"github.com/charmbracelet/log"
)

// HttpMiddleware is a middleware that logs the incoming HTTP requests
func HttpMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("request: %s %s %v", r.Method, r.URL.String(), r.ContentLength)
		ctx := r.Context()
		ctx = Context(ctx, log.With("request_method", r.Method, "request_path", r.URL.String()))

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
