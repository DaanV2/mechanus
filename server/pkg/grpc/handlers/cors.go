package grpc_handlers

import (
	"net/http"
)

type CORSOptions struct {
	Origins []string
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Credentials", "true")
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type, Cookie, X-User-Agent, X-Grpc-Web, X-Requested-With, grpc-timeout, Authorization, connect-protocol-version, connect-timeout-ms")
		h.Set("Access-Control-Expose-Headers", "Grpc-Status, Grpc-Message, Grpc-Status-Details-Bin, X-Grpc-Web, X-User-Agent, connect-protocol-version")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)

			return
		}
		next.ServeHTTP(w, r)
	})
}
