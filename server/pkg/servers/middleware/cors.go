package middleware

import "net/http"

func GRPCCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-Agent, X-Grpc-Web, X-Requested-With, grpc-timeout, Authorization, connect-protocol-version, connect-timeout-ms")
		w.Header().Set("Access-Control-Expose-Headers", "Grpc-Status, Grpc-Message, Grpc-Status-Details-Bin, X-Grpc-Web, X-User-Agent, connect-protocol-version")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)

			return
		}
		next.ServeHTTP(w, r)
	})
}
