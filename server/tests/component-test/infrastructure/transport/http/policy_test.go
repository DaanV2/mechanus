package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
	mhttp "github.com/DaanV2/mechanus/server/infrastructure/transport/http"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
)

var _ = Describe("RequirePolicy", func() {
	var (
		jtis *authentication.JTIService
		jwts *authentication.JWTService

		reached *authentication.Identity
	)

	BeforeEach(func(setupCtx SpecContext) {
		db := util_test.CreateDatabase(setupCtx)
		keys, err := authentication.NewKeyManager(storage.DBStorage[*authentication.KeyData](db))
		Expect(err).ShouldNot(HaveOccurred())
		jtis = authentication.NewJTIService(db)
		jwts = authentication.NewJWTService(jtis, keys)
		reached = nil
	})

	call := func(ctx context.Context, policy authentication.Policy, authorization string) int {
		handler := mhttp.RequirePolicy(authentication.NewIdentityResolver(jwts), policy, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reached = authentication.IdentityFromContext(r.Context())
			w.WriteHeader(http.StatusNoContent)
		}))

		req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/route", nil)
		if authorization != "" {
			req.Header.Set("Authorization", authorization)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec.Code
	}

	bearer := func(ctx context.Context, userRoles ...string) string {
		user := util_test.CreateUser()
		user.Roles = userRoles
		tokens, err := jwts.CreateSession(ctx, user, authentication.GrantPassword)
		Expect(err).ShouldNot(HaveOccurred())

		return "Bearer " + tokens.Access
	}
	device := func(ctx context.Context) string {
		tokens, err := jwts.CreateDeviceSession(ctx, "tv-1", "Table TV")
		Expect(err).ShouldNot(HaveOccurred())

		return "Device " + tokens.Access
	}

	type credential func(ctx context.Context) string
	none := func(context.Context) string { return "" }
	asUser := func(ctx context.Context) string { return bearer(ctx, roles.User.String()) }
	asAdmin := func(ctx context.Context) string { return bearer(ctx, roles.Admin.String()) }

	policies := []struct {
		policy                 authentication.Policy
		none, user, admin, dev int
	}{
		{authentication.PolicyAnonymous, http.StatusNoContent, http.StatusNoContent, http.StatusNoContent, http.StatusNoContent},
		{authentication.PolicyDevice, http.StatusUnauthorized, http.StatusNoContent, http.StatusNoContent, http.StatusNoContent},
		{authentication.PolicyUser, http.StatusUnauthorized, http.StatusNoContent, http.StatusNoContent, http.StatusForbidden},
		{authentication.PolicyRole(roles.Admin), http.StatusUnauthorized, http.StatusForbidden, http.StatusNoContent, http.StatusForbidden},
	}

	for _, p := range policies {
		Describe("policy "+p.policy.String(), func() {
			DescribeTable("answers",
				func(ctx SpecContext, cred credential, status int) {
					Expect(call(ctx, p.policy, cred(ctx))).To(Equal(status))
				},
				Entry("no credential", none, p.none),
				Entry("a user", asUser, p.user),
				Entry("an admin", asAdmin, p.admin),
				Entry("a paired device", credential(device), p.dev),
			)

			// Never downgraded to anonymous, also not on the anonymous route
			DescribeTable("rejects an invalid credential with a 401",
				func(ctx SpecContext, cred credential) {
					Expect(call(ctx, p.policy, cred(ctx))).To(Equal(http.StatusUnauthorized))
					Expect(reached).To(BeNil())
				},
				Entry("malformed bearer", credential(func(context.Context) string { return "Bearer not.a.jwt" })),
				Entry("empty bearer", credential(func(context.Context) string { return "Bearer " })),
				Entry("malformed device token", credential(func(context.Context) string { return "Device not.a.jwt" })),
				Entry("empty device token", credential(func(context.Context) string { return "Device " })),
				Entry("unknown scheme", credential(func(context.Context) string { return "Basic abc" })),
				Entry("refresh token", credential(func(ctx context.Context) string {
					tokens, err := jwts.CreateSession(ctx, util_test.CreateUser(), authentication.GrantPassword)
					Expect(err).ShouldNot(HaveOccurred())

					return "Bearer " + tokens.Refresh
				})),
				Entry("device token as bearer", credential(func(ctx context.Context) string {
					tokens, err := jwts.CreateDeviceSession(ctx, "tv-1", "Table TV")
					Expect(err).ShouldNot(HaveOccurred())

					return "Bearer " + tokens.Access
				})),
				Entry("device refresh token", credential(func(ctx context.Context) string {
					tokens, err := jwts.CreateDeviceSession(ctx, "tv-1", "Table TV")
					Expect(err).ShouldNot(HaveOccurred())

					return "Device " + tokens.Refresh
				})),
				Entry("user token as device token", credential(func(ctx context.Context) string {
					tokens, err := jwts.CreateSession(ctx, util_test.CreateUser(), authentication.GrantPassword)
					Expect(err).ShouldNot(HaveOccurred())

					return "Device " + tokens.Access
				})),
				Entry("revoked device token", credential(func(ctx context.Context) string {
					tokens, err := jwts.CreateDeviceSession(ctx, "tv-1", "Table TV")
					Expect(err).ShouldNot(HaveOccurred())
					_, err = jtis.Revoke(ctx, tokens.JTI)
					Expect(err).ShouldNot(HaveOccurred())

					return "Device " + tokens.Access
				})),
				Entry("expired device token", credential(func(ctx context.Context) string {
					// Beyond the leeway on the time claims
					jwts = jwts.WithOptions(authentication.JWTOptions{DeviceTokenDuration: -authentication.JWT_LEEWAY - time.Minute, DeviceRefreshTokenDuration: time.Hour})
					tokens, err := jwts.CreateDeviceSession(ctx, "tv-1", "Table TV")
					Expect(err).ShouldNot(HaveOccurred())

					return "Device " + tokens.Access
				})),
			)
		})
	}

	It("puts the identity on the context of the route", func(ctx SpecContext) {
		Expect(call(ctx, authentication.PolicyDevice, device(ctx))).To(Equal(http.StatusNoContent))
		Expect(reached.Kind).To(Equal(authentication.IdentityDevice))
		Expect(reached.ID).To(Equal("tv-1"))
		Expect(reached.Claims.Subject).To(Equal("tv-1"))
	})

	It("prefers a device token over the user cookie of the browser it runs on", func(ctx SpecContext) {
		handler := mhttp.RequirePolicy(authentication.NewIdentityResolver(jwts), authentication.PolicyDevice, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			reached = authentication.IdentityFromContext(r.Context())
		}))
		req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/route", nil)
		req.Header.Set("Authorization", device(ctx))
		req.AddCookie(&http.Cookie{Name: authentication.AccessTokenCookie, Value: "not-a-jwt"})
		handler.ServeHTTP(httptest.NewRecorder(), req)

		Expect(reached.Kind).To(Equal(authentication.IdentityDevice))
	})
})
