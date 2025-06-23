package grpc_handlers

import (
	"net/http"
	"slices"
	"strings"

	"github.com/DaanV2/mechanus/server/pkg/config"
	xurl "github.com/DaanV2/mechanus/server/pkg/extensions/url"
)

var (
	CorsConfig         = config.New("api.cors")
	OriginsFlag        = CorsConfig.StringArray("allowed-origins", []string{"*"}, "The origins that are allowed to be used by requesters, if empty will skip this header. Allowed strings are matched via prefix check")
	AllowLocalHostFlag = CorsConfig.Bool("allow-localhost", true, "Whenever or not as an origin, localhost are allowed")
)

type CORSConfig struct {
	AllowedOrigins []string
	AllowLocalHost bool
}

func GetCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: OriginsFlag.Value(),
		AllowLocalHost: AllowLocalHostFlag.Value(),
	}
}

type CORSHandler struct {
	AllowedOrigins []string
	AllowLocalHost bool
}

func NewCORSHandler(conf *CORSConfig) *CORSHandler {
	hand := &CORSHandler{
		AllowedOrigins: conf.AllowedOrigins,
		AllowLocalHost: conf.AllowLocalHost,
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

func Wrap(handler *CORSHandler, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Credentials", "true")
		h.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type, Cookie, X-User-Agent, X-Grpc-Web, X-Requested-With, grpc-timeout, Authorization, connect-protocol-version, connect-timeout-ms")
		h.Set("Access-Control-Expose-Headers", "Grpc-Status, Grpc-Message, Grpc-Status-Details-Bin, X-Grpc-Web, X-User-Agent, connect-protocol-version")

		if !handler.AllowOrigin(w, r) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("CORS: Origin not allowed"))
			return
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
