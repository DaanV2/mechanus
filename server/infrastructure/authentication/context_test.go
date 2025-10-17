package authentication_test

import (
	"context"

	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/golang-jwt/jwt/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Context", func() {
	ctx := context.Background()
	ctx = authentication.ContextWithJWT(ctx, &authentication.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:  "test-issuer",
			Subject: "test-subject",
		},
		User: authentication.JWTUser{
			ID:        "test-user-id",
			Name:      "Test User",
			Roles:     []string{"user"},
			Campaigns: []string{"test-campaign"},
		},
	}, true)

	DescribeTable("IsAuthenicated",
		func(role roles.Role, expected bool) {
			Expect(authentication.IsAuthenicatedWithRole(ctx, role)).To(Equal(expected))
		},
		Entry("should have viewer", roles.Viewer, true),
		Entry("should have user", roles.User, true),
		Entry("should not have operator", roles.Operator, false),
		Entry("should not have admin", roles.Admin, false),
	)
})
