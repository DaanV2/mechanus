package authentication_test

import (
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
)

var _ = Describe("JwtService", func() {
	const (
		grant = authentication.GrantPassword
		scope = authentication.ScopeAll
	)

	var (
		db         *persistence.DB
		jtiService *authentication.JTIService
		keyManager *authentication.KeyManager
		service    *authentication.JWTService
	)

	BeforeEach(func(setupCtx SpecContext) {
		var err error

		db = util_test.CreateDatabase(setupCtx)
		jtiService = authentication.NewJTIService(db)
		dbstore := storage.DBStorage[*authentication.KeyData](db)
		keyManager, err = authentication.NewKeyManager(dbstore)
		Expect(err).ShouldNot(HaveOccurred())

		service = authentication.NewJWTService(jtiService, keyManager)
	})

	claimsOf := func(ctx context.Context, token string) *authentication.JWTClaims {
		GinkgoHelper()
		t, err := service.Validate(ctx, token)
		Expect(err).ShouldNot(HaveOccurred())
		claims, ok := authentication.GetClaims(t.Claims)
		Expect(ok).To(BeTrue())

		return claims
	}

	Describe("Create", func() {
		It("should create a JWT in a new session (JTI) every time", func(ctx SpecContext) {
			user := util_test.CreateUser()

			token1, err := service.Create(ctx, user, grant, scope)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token1).ToNot(BeEmpty())

			token2, err := service.Create(ctx, user, grant, scope)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token2).ToNot(BeEmpty())

			claims1, claims2 := claimsOf(ctx, token1), claimsOf(ctx, token2)
			Expect(claims1.ID).ToNot(Equal(claims2.ID))
			Expect(claims1.User.ID).To(Equal(user.ID))
		})
	})

	Describe("CreateSession", func() {
		It("should return an access and refresh token of the same session", func(ctx SpecContext) {
			user := util_test.CreateUser()

			tokens, err := service.CreateSession(ctx, user, grant)
			Expect(err).ShouldNot(HaveOccurred())

			access, refresh := claimsOf(ctx, tokens.Access), claimsOf(ctx, tokens.Refresh)
			Expect(access.ID).To(Equal(refresh.ID))
			Expect(access.Scope).To(Equal(authentication.ScopeAll))
			Expect(refresh.Scope).To(Equal(authentication.ScopeRefresh))
			Expect(refresh.ExpiresAt.Time).To(BeTemporally(">", access.ExpiresAt.Time))
		})

		It("RefreshSession keeps the session and picks up new roles", func(ctx SpecContext) {
			user := util_test.CreateUser()
			tokens, err := service.CreateSession(ctx, user, grant)
			Expect(err).ShouldNot(HaveOccurred())
			jti := claimsOf(ctx, tokens.Refresh).ID

			user.Roles = []string{"admin"}
			refreshed, err := service.RefreshSession(ctx, user, jti)
			Expect(err).ShouldNot(HaveOccurred())

			access := claimsOf(ctx, refreshed.Access)
			Expect(access.ID).To(Equal(jti))
			Expect(access.Grant).To(Equal(authentication.GrantRefresh))
			Expect(access.User.Roles).To(ConsistOf("admin"))
		})
	})

	Describe("Validate", func() {
		It("should validate a valid JWT and return claims", func(ctx SpecContext) {
			user := util_test.CreateUser()
			token, err := service.Create(ctx, user, grant, scope)
			Expect(err).ShouldNot(HaveOccurred())

			claims := claimsOf(ctx, token)
			Expect(claims.User.ID).To(Equal(user.ID))
			Expect(claims.Grant).To(Equal(grant))
			Expect(claims.Scope).To(Equal(scope))
		})

		It("should fail validation if the JTI is revoked", func(ctx SpecContext) {
			user := util_test.CreateUser()
			token, err := service.Create(ctx, user, grant, scope)
			Expect(err).ShouldNot(HaveOccurred())

			_, err = jtiService.Revoke(ctx, claimsOf(ctx, token).ID)
			Expect(err).ShouldNot(HaveOccurred())

			_, err = service.Validate(ctx, token)
			Expect(err).To(MatchError(authentication.ErrJTIRevoked))
		})

		It("should fail an expired token", func(ctx SpecContext) {
			service.Options().TokenDuration = -time.Hour
			token, err := service.Create(ctx, util_test.CreateUser(), grant, scope)
			Expect(err).ShouldNot(HaveOccurred())

			_, err = service.Validate(ctx, token)
			Expect(err).To(MatchError(jwt.ErrTokenExpired))
		})

		It("should accept a token that expired within the leeway", func(ctx SpecContext) {
			service.Options().TokenDuration = -time.Minute
			token, err := service.Create(ctx, util_test.CreateUser(), grant, scope)
			Expect(err).ShouldNot(HaveOccurred())

			_, err = service.Validate(ctx, token)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should fail a tampered signature", func(ctx SpecContext) {
			token, err := service.Create(ctx, util_test.CreateUser(), grant, scope)
			Expect(err).ShouldNot(HaveOccurred())
			i := strings.LastIndex(token, ".") + 10
			replacement := byte('A')
			if token[i] == 'A' {
				replacement = 'B'
			}
			token = token[:i] + string(replacement) + token[i+1:]

			_, err = service.Validate(ctx, token)
			Expect(err).To(MatchError(jwt.ErrTokenSignatureInvalid))
		})

		Describe("forged claims", func() {
			var user = util_test.CreateUser()

			// sign builds a token for a fresh session of the user, lets modify change it, then signs it with the server key
			sign := func(ctx context.Context, modify func(claims *authentication.JWTClaims, token *jwt.Token)) string {
				GinkgoHelper()
				jti, err := jtiService.Create(ctx, user.ID)
				Expect(err).ShouldNot(HaveOccurred())
				key, err := keyManager.GetSigningKey(ctx)
				Expect(err).ShouldNot(HaveOccurred())

				now := time.Now()
				claims := &authentication.JWTClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						ID:        jti.ID,
						Issuer:    authentication.JWT_ISSUER,
						Audience:  jwt.ClaimStrings{authentication.JWT_AUDIENCE},
						Subject:   user.ID,
						IssuedAt:  jwt.NewNumericDate(now),
						NotBefore: jwt.NewNumericDate(now),
						ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
					},
					User:  authentication.JWTUser{ID: user.ID, Name: user.Username, Roles: user.Roles},
					Grant: grant,
					Scope: scope,
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
				token.Header["kid"] = key.GetID()
				modify(claims, token)

				result, err := token.SignedString(key.Private())
				Expect(err).ShouldNot(HaveOccurred())

				return result
			}

			It("accepts an unmodified token", func(ctx SpecContext) {
				token := sign(ctx, func(*authentication.JWTClaims, *jwt.Token) {})

				_, err := service.Validate(ctx, token)
				Expect(err).ShouldNot(HaveOccurred())
			})

			DescribeTable("rejects",
				func(ctx SpecContext, modify func(claims *authentication.JWTClaims, token *jwt.Token), expected error) {
					token := sign(ctx, modify)

					_, err := service.Validate(ctx, token)
					Expect(err).To(HaveOccurred())
					if expected != nil {
						Expect(err).To(MatchError(expected))
					}
				},
				Entry("a not-yet-valid token", func(c *authentication.JWTClaims, _ *jwt.Token) {
					c.NotBefore = jwt.NewNumericDate(time.Now().Add(time.Hour))
				}, jwt.ErrTokenNotValidYet),
				Entry("a missing expiry", func(c *authentication.JWTClaims, _ *jwt.Token) {
					c.ExpiresAt = nil
				}, jwt.ErrTokenRequiredClaimMissing),
				Entry("a wrong issuer", func(c *authentication.JWTClaims, _ *jwt.Token) {
					c.Issuer = "someone-else"
				}, jwt.ErrTokenInvalidIssuer),
				Entry("a wrong audience", func(c *authentication.JWTClaims, _ *jwt.Token) {
					c.Audience = jwt.ClaimStrings{"someone-else"}
				}, jwt.ErrTokenInvalidAudience),
				Entry("a missing kid", func(_ *authentication.JWTClaims, t *jwt.Token) {
					delete(t.Header, "kid")
				}, authentication.ErrKIDMissing),
				Entry("a non-string kid", func(_ *authentication.JWTClaims, t *jwt.Token) {
					t.Header["kid"] = 42
				}, authentication.ErrKIDNotString),
				Entry("an unknown kid", func(_ *authentication.JWTClaims, t *jwt.Token) {
					t.Header["kid"] = "does-not-exist"
				}, nil),
				Entry("a JTI that belongs to another user", func(c *authentication.JWTClaims, _ *jwt.Token) {
					c.User.ID = util_test.CreateUserID()
				}, authentication.ErrJTIWrongUser),
				Entry("an unknown JTI", func(c *authentication.JWTClaims, _ *jwt.Token) {
					c.ID = "does-not-exist"
				}, nil),
			)
		})
	})

	Describe("Cleanup", func() {
		It("removes revoked sessions and keeps active ones", func(ctx SpecContext) {
			user := util_test.CreateUser()
			active, err := service.Create(ctx, user, grant, scope)
			Expect(err).ShouldNot(HaveOccurred())
			revoked, err := service.Create(ctx, user, grant, scope)
			Expect(err).ShouldNot(HaveOccurred())
			revokedJTI := claimsOf(ctx, revoked).ID
			_, err = jtiService.Revoke(ctx, revokedJTI)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(service.Cleanup(ctx)).To(Succeed())

			_, err = jtiService.Get(ctx, revokedJTI)
			Expect(err).To(HaveOccurred())
			_, err = service.Validate(ctx, active)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
