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
	// WebsocketDeviceSubprotocolPrefix prefixes a device token when a browser passes it as a WebSocket subprotocol.
	WebsocketDeviceSubprotocolPrefix = "mechanus.device."

	// DeviceIDHeader is the header a device reports its stable device ID in.
	DeviceIDHeader = "X-Device-Id"
	// DeviceTypeHeader is the header a device reports the hardware it runs on in.
	DeviceTypeHeader = "X-Device-Type"

	// DeviceTokenType is the prefix of a device token in the Authorization header.
	DeviceTokenType = "Device"

	bearerPrefix = "Bearer "
	devicePrefix = DeviceTokenType + " "
)

// DeviceTokenFromHeaders returns the device token (without the `Device ` prefix) carried by a request, and whether
// there was one. Sources are checked in order: the Authorization header and a `mechanus.device.<jwt>` WebSocket subprotocol.
func DeviceTokenFromHeaders(headers http.Header) (string, bool) {
	for _, auth := range headers.Values("Authorization") {
		if token, ok := strings.CutPrefix(auth, devicePrefix); ok {
			return token, true
		}
	}

	for _, protocols := range headers.Values("Sec-WebSocket-Protocol") {
		for protocol := range strings.SplitSeq(protocols, ",") {
			if token, ok := strings.CutPrefix(strings.TrimSpace(protocol), WebsocketDeviceSubprotocolPrefix); ok {
				return token, true
			}
		}
	}

	return "", false
}

// hasUnknownAuthorization reports whether the request has an Authorization header that is neither a non-empty
// bearer token nor a device token.
func hasUnknownAuthorization(headers http.Header) bool {
	for _, auth := range headers.Values("Authorization") {
		token, ok := strings.CutPrefix(auth, bearerPrefix)
		if (!ok || token == "") && !strings.HasPrefix(auth, devicePrefix) {
			return true
		}
	}

	return false
}

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
