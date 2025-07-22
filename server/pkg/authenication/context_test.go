package authenication_test

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/pkg/authenication"
	"github.com/DaanV2/mechanus/server/pkg/authenication/roles"
)

var _ = Describe("Context", func() {
	ctx := context.Background()
	ctx = authenication.ContextWithJWT(ctx, &authenication.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:  "test-issuer",
			Subject: "test-subject",
		},
		User: authenication.JWTUser{
			ID:        "test-user-id",
			Name:      "Test User",
			Roles:     []string{"user"},
			Campaigns: []string{"test-campaign"},
		},
	}, true)

	DescribeTable("IsAuthenicated",
		func(role roles.Role, expected bool) {
			Expect(authenication.IsAuthenicatedWithRole(ctx, role)).To(Equal(expected))
		},
		Entry("should have viewer", roles.Viewer, true),
		Entry("should have user", roles.User, true),
		Entry("should not have operator", roles.Operator, false),
		Entry("should not have admin", roles.Admin, false),
	)
})
