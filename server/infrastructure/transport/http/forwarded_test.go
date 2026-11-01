package http_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mhttp "github.com/DaanV2/mechanus/server/infrastructure/transport/http"
)

var _ = Describe("ForwardedHeaders", func() {
	// serve passes a request with the given headers through the middleware, and returns the request the next handler got
	serve := func(headers map[string]string) *http.Request {
		var seen *http.Request
		handler := mhttp.NewForwardedHeaders(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { seen = r }))

		req := httptest.NewRequest(http.MethodGet, "http://internal:8080/api", http.NoBody)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		handler.ServeHTTP(httptest.NewRecorder(), req)

		return seen
	}

	It("applies the scheme, host and client address of the proxy", func() {
		r := serve(map[string]string{
			mhttp.ForwardedProtoHeader: "HTTPS",
			mhttp.ForwardedHostHeader:  "mechanus.example.com",
			mhttp.ForwardedForHeader:   "203.0.113.7, 10.0.0.2",
		})

		Expect(r.URL.Scheme).To(Equal("https"))
		Expect(r.Host).To(Equal("mechanus.example.com"))
		Expect(r.RemoteAddr).To(Equal("203.0.113.7:0"))
	})

	It("leaves the request alone without forwarded headers", func() {
		r := serve(nil)

		Expect(r.Host).To(Equal("internal:8080"))
		Expect(r.RemoteAddr).To(Equal("192.0.2.1:1234"))
	})

	It("ignores unknown schemes and invalid addresses", func() {
		r := serve(map[string]string{
			mhttp.ForwardedProtoHeader: "gopher",
			mhttp.ForwardedForHeader:   "not-an-ip",
		})

		Expect(r.URL.Scheme).To(Equal("http"))
		Expect(r.RemoteAddr).To(Equal("192.0.2.1:1234"))
	})
})
