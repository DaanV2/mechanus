package http

import (
	"net"
	"net/http"
	"strings"
)

// Headers a reverse proxy sets to describe the original request.
const (
	ForwardedProtoHeader = "X-Forwarded-Proto"
	ForwardedHostHeader  = "X-Forwarded-Host"
	ForwardedForHeader   = "X-Forwarded-For"
)

var _ http.Handler = &ForwardedHeaders{}

// ForwardedHeaders applies the X-Forwarded-* headers of a trusted reverse proxy to the request, so the rest
// of the server (logs, traces and the WebSocket origin check, which compares the Origin with the host) sees the
// request as the client made it: its scheme, host and client address. Only use it when every request goes
// through the proxy, otherwise clients can spoof these headers.
type ForwardedHeaders struct {
	next http.Handler
}

// NewForwardedHeaders wraps next with [ForwardedHeaders].
func NewForwardedHeaders(next http.Handler) *ForwardedHeaders {
	return &ForwardedHeaders{next}
}

// ServeHTTP implements http.Handler.
func (f *ForwardedHeaders) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proto := strings.ToLower(firstValue(r.Header.Get(ForwardedProtoHeader)))
	host := firstValue(r.Header.Get(ForwardedHostHeader))
	client := firstValue(r.Header.Get(ForwardedForHeader))

	if proto == "http" || proto == "https" || host != "" || client != "" {
		r = r.Clone(r.Context())
		if proto == "http" || proto == "https" {
			r.URL.Scheme = proto
		}
		if host != "" {
			r.Host = host
		}
		if ip := net.ParseIP(client); ip != nil {
			r.RemoteAddr = net.JoinHostPort(ip.String(), "0")
		}
	}

	f.next.ServeHTTP(w, r)
}

// firstValue returns the first entry of a comma separated header, the one the client side proxy added.
func firstValue(header string) string {
	value, _, _ := strings.Cut(header, ",")

	return strings.TrimSpace(value)
}
