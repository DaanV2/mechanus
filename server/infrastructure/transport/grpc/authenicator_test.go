package grpc

import (
	"net/http"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("getJwtValue", func() {
	const jwt = "eyJhbGciOiJSUzUxMiJ9.eyJzdWIiOiIxIn0.c2lnbmF0dXJl"

	It("returns the Authorization bearer header", func() {
		headers := http.Header{}
		headers.Set("Authorization", "Bearer "+jwt)

		Expect(getJwtValue(headers)).To(Equal("Bearer " + jwt))
	})

	It("prefers the Authorization header over the cookie", func() {
		headers := http.Header{}
		headers.Set("Authorization", "Bearer header")
		headers.Set("Cookie", "access-token=Bearer%20cookie")

		Expect(getJwtValue(headers)).To(Equal("Bearer header"))
	})

	It("decodes the URI-encoded access-token cookie set by the client", func() {
		headers := http.Header{}
		headers.Set("Cookie", "theme=dark; access-token="+url.PathEscape("Bearer "+jwt))

		Expect(getJwtValue(headers)).To(Equal("Bearer " + jwt))
	})

	It("ignores cookies without a bearer value", func() {
		headers := http.Header{}
		headers.Set("Cookie", "access-token="+jwt)

		Expect(getJwtValue(headers)).To(BeEmpty())
	})

	It("ignores malformed escapes", func() {
		headers := http.Header{}
		headers.Set("Cookie", "access-token=Bearer%2")

		Expect(getJwtValue(headers)).To(BeEmpty())
	})

	It("returns empty when nothing is present", func() {
		Expect(getJwtValue(http.Header{})).To(BeEmpty())
	})
})
