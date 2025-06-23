package xurl

import "strings"

func IsLocalHostOrigin(origin string) bool {
	switch true {
	case len(origin) == 0:
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
