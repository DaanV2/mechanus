package authentication_test

import (
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JWTClaims", func() {
	DescribeTable("HasScope",
		func(tokenScope, requested string, expected bool) {
			claims := &authentication.JWTClaims{Scope: tokenScope}
			Expect(claims.HasScope(requested)).To(Equal(expected))
		},
		Entry("all scope allows all", authentication.ScopeAll, authentication.ScopeAll, true),
		Entry("all scope allows refresh", authentication.ScopeAll, authentication.ScopeRefresh, true),
		Entry("refresh scope allows refresh", authentication.ScopeRefresh, authentication.ScopeRefresh, true),
		Entry("refresh scope does not allow all", authentication.ScopeRefresh, authentication.ScopeAll, false),
		Entry("empty scope allows nothing", "", authentication.ScopeRefresh, false),
	)
})
