package grpc_test

import (
	"net/http"
	"net/url"

	"github.com/DaanV2/mechanus/server/infrastructure/transport/grpc"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("getJwtValue", func() {
	const jwt = "eyJhbGciOiJSUzUxMiJ9.eyJzdWIiOiIxIn0.c2lnbmF0dXJl"

	It("returns the Authorization bearer header", func() {
		headers := http.Header{}
		headers.Set("Authorization", "Bearer "+jwt)

		Expect(grpc.GetJwtValue(headers)).To(Equal("Bearer " + jwt))
	})

	It("prefers the Authorization header over the cookie", func() {
		headers := http.Header{}
		headers.Set("Authorization", "Bearer header")
		headers.Set("Cookie", "access-token=Bearer%20cookie")

		Expect(grpc.GetJwtValue(headers)).To(Equal("Bearer header"))
	})

	It("decodes the URI-encoded access-token cookie set by the client", func() {
		headers := http.Header{}
		headers.Set("Cookie", "theme=dark; access-token="+url.PathEscape("Bearer "+jwt))

		Expect(grpc.GetJwtValue(headers)).To(Equal("Bearer " + jwt))
	})

	It("ignores cookies without a bearer value", func() {
		headers := http.Header{}
		headers.Set("Cookie", "access-token="+jwt)

		Expect(grpc.GetJwtValue(headers)).To(BeEmpty())
	})

	It("ignores malformed escapes", func() {
		headers := http.Header{}
		headers.Set("Cookie", "access-token=Bearer%2")

		Expect(grpc.GetJwtValue(headers)).To(BeEmpty())
	})

	It("returns empty when nothing is present", func() {
		Expect(grpc.GetJwtValue(http.Header{})).To(BeEmpty())
	})
})
