package cors

import (
	"net/http"
	"slices"
	"strings"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xurl"
	"github.com/charmbracelet/log"
)

var _ http.Handler = &CORSHandler{}

type CORSHandler struct {
	AllowedOrigins []string
	AllowLocalHost bool
	Next           http.Handler
}

func NewCORSHandler(conf *CORSConfig, next http.Handler) *CORSHandler {
	hand := &CORSHandler{
		AllowedOrigins: conf.AllowedOrigins,
		AllowLocalHost: conf.AllowLocalHost,
		Next:           next,
	}

	if slices.Contains(hand.AllowedOrigins, "*") {
		hand.AllowedOrigins = []string{"*"}
		hand.AllowLocalHost = true
	}

	return hand
}

func (hand *CORSHandler) AllowOrigin(w http.ResponseWriter, r *http.Request) bool {
	header := w.Header()
	origin := r.Header.Get("Origin")

	if hand.AllowLocalHost && xurl.IsLocalHostOrigin(origin) {
		header.Set("Access-Control-Allow-Origin", origin)

		return true
	}

	switch len(hand.AllowedOrigins) {
	case 0:
	case 1:
		header.Set("Access-Control-Allow-Origin", hand.AllowedOrigins[0])

		return origin == hand.AllowedOrigins[0] || hand.AllowedOrigins[0] == "*"
	default:
		for _, o := range hand.AllowedOrigins {
			if o == "*" {
				header.Set("Access-Control-Allow-Origin", origin)

				return true
			}
			if strings.HasPrefix(origin, o) {
				header.Set("Access-Control-Allow-Origin", origin)

				return true
			}
		}
	}

	return false
}

// ServeHTTP implements http.Handler.
func (hand *CORSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	h.Set("Access-Control-Allow-Credentials", "true")
	h.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	h.Set("Access-Control-Allow-Headers", "Content-Type, Cookie, X-User-Agent, X-Grpc-Web, X-Requested-With, grpc-timeout, Authorization, connect-protocol-version, connect-timeout-ms")
	h.Set("Access-Control-Expose-Headers", "Grpc-Status, Grpc-Message, Grpc-Status-Details-Bin, X-Grpc-Web, X-User-Agent, connect-protocol-version")

	if !hand.AllowOrigin(w, r) {
		w.WriteHeader(http.StatusForbidden)
		_, err := w.Write([]byte("CORS: Origin not allowed"))
		if err != nil {
			log.WithPrefix("cors").Error("error during writing 403: CORS origin not allowed", "error", err)
		}

		return
	}

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)

		return
	}

	hand.Next.ServeHTTP(w, r)
}
