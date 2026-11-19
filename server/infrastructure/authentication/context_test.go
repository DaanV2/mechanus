package authentication_test

import (
	"context"

	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Context", func() {
	authenticated := func(ctx context.Context) context.Context {
		return authentication.ContextWithJWT(ctx, &authentication.JWTClaims{
			Issuer:  "test-issuer",
			Subject: "test-subject",
			User: authentication.JWTUser{
				ID:        "test-user-id",
				Name:      "Test User",
				Roles:     []string{"user"},
				Campaigns: []string{"test-campaign"},
			},
		}, true)
	}

	DescribeTable("IsAuthenicated",
		func(ctx SpecContext, role roles.Role, expected bool) {
			Expect(authentication.IsAuthenicatedWithRole(authenticated(ctx), role)).To(Equal(expected))
		},
		Entry("should have viewer", roles.Viewer, true),
		Entry("should have user", roles.User, true),
		Entry("should not have operator", roles.Operator, false),
		Entry("should not have admin", roles.Admin, false),
	)
})
