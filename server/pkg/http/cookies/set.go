package cookies

import "net/http"

type HeaderContainer interface {
	Header() http.Header
}

func Set(resp HeaderContainer, cookie *http.Cookie) {
	if v := cookie.String(); v != "" {
		resp.Header().Add("Set-Cookie", v)
	}
}
