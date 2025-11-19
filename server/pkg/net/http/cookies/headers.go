package cookies

import (
	"net/http"
	"strings"
)

// Cookie represents a simple name-value pair for a cookie.
type Cookie struct {
	Name, Value string
}

// HeaderContainer is an interface for types that provide HTTP headers.
type HeaderContainer interface {
	Header() http.Header
}

// Set adds a Set-Cookie header to the response.
func Set(resp HeaderContainer, cookie *http.Cookie) {
	if v := cookie.String(); v != "" {
		resp.Header().Add("Set-Cookie", v)
	}
}

// SetCookies adds multiple cookies to the response with appropriate settings.
func SetCookies(resp, req HeaderContainer, cs ...*Cookie) {
	domain := extractDomain(req)
	for _, c := range cs {
		cookie := &http.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Path:     "/",
			Domain:   domain,
			Secure:   true,
			HttpOnly: false,
			SameSite: http.SameSiteLaxMode,
		}
		Set(resp, cookie)
	}
}

func extractDomain(req HeaderContainer) string {
	origin := req.Header().Get("Origin")
	if idx := strings.Index(origin, "://"); idx != -1 {
		origin = origin[idx+3:]
	}
	if idx := strings.Index(origin, "/"); idx != -1 {
		origin = origin[:idx]
	}
	// Remove port if present
	if idx := strings.Index(origin, ":"); idx != -1 {
		origin = origin[:idx]
	}

	return origin
}
