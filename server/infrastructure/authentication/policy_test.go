package authentication_test

import (
	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Policy", func() {
	user := func(userRoles ...string) *authentication.Identity {
		claims := &authentication.JWTClaims{User: authentication.JWTUser{ID: "user-1", Roles: userRoles}}

		return &authentication.Identity{Kind: authentication.IdentityUser, ID: "user-1", Claims: claims}
	}
	device := func() *authentication.Identity {
		claims := &authentication.JWTClaims{User: authentication.JWTUser{ID: "tv-1", Roles: []string{roles.Device.String()}}}

		return &authentication.Identity{Kind: authentication.IdentityDevice, ID: "tv-1", Claims: claims}
	}

	DescribeTable("allows",
		func(policy authentication.Policy, identity func() *authentication.Identity) {
			Expect(policy.Authorize(identity())).To(Succeed())
		},
		Entry("anonymous: no identity", authentication.PolicyAnonymous, authentication.Anonymous),
		Entry("anonymous: a device", authentication.PolicyAnonymous, device),
		Entry("anonymous: a user", authentication.PolicyAnonymous, func() *authentication.Identity { return user("user") }),
		Entry("device: a device", authentication.PolicyDevice, device),
		Entry("device: a user", authentication.PolicyDevice, func() *authentication.Identity { return user("user") }),
		Entry("user: a user", authentication.PolicyUser, func() *authentication.Identity { return user("user") }),
		Entry("role: a user with the role", authentication.PolicyRole(roles.Admin), func() *authentication.Identity { return user("admin") }),
		Entry("role: a user with an inheriting role", authentication.PolicyRole(roles.User), func() *authentication.Identity { return user("admin") }),
	)

	DescribeTable("answers that an identity is needed",
		func(policy authentication.Policy) {
			Expect(policy.Authorize(authentication.Anonymous())).To(MatchError(authentication.ErrIdentityRequired))
		},
		Entry("device", authentication.PolicyDevice),
		Entry("user", authentication.PolicyUser),
		Entry("role", authentication.PolicyRole(roles.Admin)),
	)

	DescribeTable("answers that the identity isn't allowed",
		func(policy authentication.Policy, identity func() *authentication.Identity) {
			Expect(policy.Authorize(identity())).To(MatchError(authentication.ErrNotAllowed))
		},
		Entry("user: a device", authentication.PolicyUser, device),
		Entry("role: a device", authentication.PolicyRole(roles.Device), device),
		Entry("role: a user without the role", authentication.PolicyRole(roles.Admin), func() *authentication.Identity { return user("user") }),
		Entry("role: a user without roles", authentication.PolicyRole(roles.User), func() *authentication.Identity { return user() }),
	)

	DescribeTable("is named after what it accepts",
		func(policy authentication.Policy, name string) {
			Expect(policy.String()).To(Equal(name))
		},
		Entry("anonymous", authentication.PolicyAnonymous, "anonymous"),
		Entry("device", authentication.PolicyDevice, "device"),
		Entry("user", authentication.PolicyUser, "user"),
		Entry("role", authentication.PolicyRole(roles.Admin), "role:admin"),
	)
})

var _ = Describe("IdentityFromContext", func() {
	It("is anonymous without an identity on the context", func(ctx SpecContext) {
		Expect(authentication.IdentityFromContext(ctx).Kind).To(Equal(authentication.IdentityAnonymous))
	})

	It("returns the identity on the context", func(ctx SpecContext) {
		identity := &authentication.Identity{Kind: authentication.IdentityDevice, ID: "tv-1"}
		Expect(authentication.IdentityFromContext(authentication.ContextWithIdentity(ctx, identity))).To(BeIdenticalTo(identity))
	})
})
