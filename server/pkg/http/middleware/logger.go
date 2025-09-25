package http_middleware

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/charmbracelet/log"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("request: %s %s %v", r.Method, r.URL.String(), r.ContentLength)

		r = r.WithContext(logging.Context(r.Context(), log.With("request_method", r.Method, "request_path", r.URL.String())))

		next.ServeHTTP(w, r)
	})
}
