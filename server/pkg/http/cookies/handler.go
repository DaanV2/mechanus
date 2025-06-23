package cookies

import (
	"net/http"
	"strings"
)

type Cookie struct {
	Name, Value string
}

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
