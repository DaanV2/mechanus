package xurl_test

import (
	xurl "github.com/DaanV2/mechanus/server/pkg/extensions/xurl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IsLocalHostOrigin", func() {
	DescribeTable("returns expected result",
		func(origin string, expected bool) {
			Expect(xurl.IsLocalHostOrigin(origin)).To(Equal(expected))
		},
		Entry("http://localhost", "http://localhost", true),
		Entry("http://127.0.0.1", "http://127.0.0.1", true),
		Entry("https://localhost", "https://localhost", true),
		Entry("https://127.0.0.1", "https://127.0.0.1", true),
		Entry("http://example.com", "http://example.com", false),
		Entry("https://192.168.1.1", "https://192.168.1.1", false),
		Entry("empty string", "", false),
	)
})
