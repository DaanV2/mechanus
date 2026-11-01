package authentication

import (
	"net/http"
	"net/url"
	"strings"
)

const (
	// AccessTokenCookie is the name of the cookie the browser client stores its access token in.
	// Its value is `encodeURIComponent("Bearer <jwt>")`.
	AccessTokenCookie = "access-token"

	// WebsocketSubprotocol is the subprotocol the screen WebSocket negotiates with a client.
	WebsocketSubprotocol = "mechanus.v1"
	// WebsocketBearerSubprotocolPrefix prefixes the JWT when a browser passes it as a WebSocket subprotocol,
	// since browsers can't set an Authorization header on `new WebSocket(...)`.
	WebsocketBearerSubprotocolPrefix = "mechanus.bearer."

	bearerPrefix = "Bearer "
)

// BearerFromHeaders returns the bearer JWT (without the `Bearer ` prefix) carried by a request, or "" if there is none.
// Sources are checked in order: the Authorization header, a `mechanus.bearer.<jwt>` WebSocket subprotocol,
// and the access token cookie.
func BearerFromHeaders(headers http.Header) string {
	for _, auth := range headers.Values("Authorization") {
		if token, ok := strings.CutPrefix(auth, bearerPrefix); ok {
			return token
		}
	}

	for _, protocols := range headers.Values("Sec-WebSocket-Protocol") {
		for protocol := range strings.SplitSeq(protocols, ",") {
			if token, ok := strings.CutPrefix(strings.TrimSpace(protocol), WebsocketBearerSubprotocolPrefix); ok {
				return token
			}
		}
	}

	for _, cookieData := range headers.Values("Cookie") {
		cookies, err := http.ParseCookie(cookieData)
		if err != nil {
			continue
		}

		for _, cookie := range cookies {
			if cookie.Name != AccessTokenCookie {
				continue
			}

			// The client URL-encodes the cookie value, so `Bearer <jwt>` arrives as `Bearer%20<jwt>`
			value, err := url.PathUnescape(cookie.Value)
			if err != nil {
				continue
			}
			if token, ok := strings.CutPrefix(value, bearerPrefix); ok {
				return token
			}
		}
	}

	return ""
}
