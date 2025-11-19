package xurl

import "strings"

// IsLocalHostOrigin checks if the given origin URL is a localhost URL (http or https).
func IsLocalHostOrigin(origin string) bool {
	switch {
	case origin == "":
		return false
	case strings.HasPrefix(origin, "http://localhost"):
	case strings.HasPrefix(origin, "http://127.0.0.1"):
	case strings.HasPrefix(origin, "https://localhost"):
	case strings.HasPrefix(origin, "https://127.0.0.1"):
	default:
		return false
	}

	return true
}
