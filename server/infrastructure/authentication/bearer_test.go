package authentication_test

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BearerFromHeaders", func() {
	DescribeTable("finds the bearer token",
		func(headers map[string][]string, expected string) {
			Expect(authentication.BearerFromHeaders(http.Header(headers))).To(Equal(expected))
		},
		Entry("no headers", map[string][]string{}, ""),
		Entry("authorization header", map[string][]string{"Authorization": {"Bearer abc.def.ghi"}}, "abc.def.ghi"),
		Entry("non-bearer authorization header", map[string][]string{"Authorization": {"Device abc"}}, ""),
		Entry("websocket subprotocol",
			map[string][]string{"Sec-Websocket-Protocol": {"mechanus.v1, mechanus.bearer.abc.def.ghi"}}, "abc.def.ghi"),
		Entry("websocket subprotocol without bearer", map[string][]string{"Sec-Websocket-Protocol": {"mechanus.v1"}}, ""),
		Entry("url-encoded cookie, as the client stores it",
			map[string][]string{"Cookie": {"other=1; access-token=Bearer%20abc.def.ghi"}}, "abc.def.ghi"),
		Entry("plain cookie", map[string][]string{"Cookie": {`access-token="Bearer abc.def.ghi"`}}, "abc.def.ghi"),
		Entry("cookie without bearer prefix", map[string][]string{"Cookie": {"access-token=abc.def.ghi"}}, ""),
		Entry("other cookie", map[string][]string{"Cookie": {"refresh-token=Bearer%20abc"}}, ""),
		Entry("authorization header wins over subprotocol and cookie",
			map[string][]string{
				"Authorization":          {"Bearer header"},
				"Sec-Websocket-Protocol": {"mechanus.v1, mechanus.bearer.protocol"},
				"Cookie":                 {"access-token=Bearer%20cookie"},
			}, "header"),
		Entry("subprotocol wins over cookie",
			map[string][]string{
				"Sec-Websocket-Protocol": {"mechanus.v1, mechanus.bearer.protocol"},
				"Cookie":                 {"access-token=Bearer%20cookie"},
			}, "protocol"),
	)
})
