package grpc_test

import (
	"connectrpc.com/connect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/grpc"
	usersv1 "github.com/DaanV2/mechanus/server/proto/users/v1"
)

var _ = Describe("LoginServiceHandler", func() {
	const password = "correct horse battery staple"

	var (
		env     *authEnv
		handler *grpc.LoginServiceHandler
		user    *models.User
	)

	BeforeEach(func(ctx SpecContext) {
		env = newAuthEnv(ctx)
		handler = grpc.NewLoginServiceHandler(env.users, env.jwts, env.jtis)
		user = env.createUser(ctx, password, "user")
	})

	login := func(ctx SpecContext, username, password string) (*usersv1.LoginResponse, error) {
		resp, err := handler.Login(ctx, connect.NewRequest(&usersv1.LoginRequest{Username: username, Password: password}))
		if err != nil {
			return nil, err
		}

		return resp.Msg, nil
	}

	Describe("Login", func() {
		It("returns an access and refresh token of a new session", func(ctx SpecContext) {
			resp, err := login(ctx, user.Username, password)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.GetType()).To(Equal("Bearer"))

			access, refresh := env.claims(ctx, resp.GetToken()), env.claims(ctx, resp.GetRefreshToken())
			Expect(access.Scope).To(Equal(authentication.ScopeAll))
			Expect(refresh.Scope).To(Equal(authentication.ScopeRefresh))
			Expect(access.ID).To(Equal(refresh.ID))
			Expect(access.User.ID).To(Equal(user.ID))

			// A second login is a separate session
			again, err := login(ctx, user.Username, password)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(env.claims(ctx, again.GetToken()).ID).ToNot(Equal(access.ID))
		})

		DescribeTable("rejects invalid credentials as unauthenticated",
			func(ctx SpecContext, username, pass string) {
				if username == "<user>" {
					username = user.Username
				}

				_, err := login(ctx, username, pass)
				Expect(err).To(HaveConnectCode(connect.CodeUnauthenticated))
			},
			Entry("wrong password", "<user>", "wrong"),
			Entry("unknown user", "nobody", password),
			Entry("empty username", "", password),
			Entry("empty password", "<user>", ""),
		)

		It("locks a username out after too many failed attempts", func(ctx SpecContext) {
			for range 5 {
				_, err := login(ctx, user.Username, "wrong")
				Expect(err).To(HaveConnectCode(connect.CodeUnauthenticated))
			}

			_, err := login(ctx, user.Username, password)
			Expect(err).To(HaveConnectCode(connect.CodeResourceExhausted))
		})
	})

	Describe("Refresh", func() {
		refresh := func(ctx SpecContext, token string) (*usersv1.RefreshTokenResponse, error) {
			resp, err := handler.Refresh(ctx, connect.NewRequest(&usersv1.RefreshTokenRequest{Token: token}))
			if err != nil {
				return nil, err
			}

			return resp.Msg, nil
		}

		It("exchanges a refresh token for new tokens in the same session", func(ctx SpecContext) {
			loggedIn, err := login(ctx, user.Username, password)
			Expect(err).ShouldNot(HaveOccurred())

			resp, err := refresh(ctx, loggedIn.GetRefreshToken())
			Expect(err).ShouldNot(HaveOccurred())

			access := env.claims(ctx, resp.GetToken())
			Expect(access.ID).To(Equal(env.claims(ctx, loggedIn.GetToken()).ID))
			Expect(access.Scope).To(Equal(authentication.ScopeAll))
			Expect(access.Grant).To(Equal(authentication.GrantRefresh))
			Expect(env.claims(ctx, resp.GetRefreshToken()).Scope).To(Equal(authentication.ScopeRefresh))
		})

		It("does not refresh an access token", func(ctx SpecContext) {
			loggedIn, err := login(ctx, user.Username, password)
			Expect(err).ShouldNot(HaveOccurred())

			_, err = refresh(ctx, loggedIn.GetToken())
			Expect(err).To(HaveConnectCode(connect.CodePermissionDenied))
		})

		It("does not refresh a revoked session", func(ctx SpecContext) {
			loggedIn, err := login(ctx, user.Username, password)
			Expect(err).ShouldNot(HaveOccurred())
			_, err = env.jtis.Revoke(ctx, env.claims(ctx, loggedIn.GetRefreshToken()).ID)
			Expect(err).ShouldNot(HaveOccurred())

			_, err = refresh(ctx, loggedIn.GetRefreshToken())
			Expect(err).To(HaveConnectCode(connect.CodeUnauthenticated))
		})

		It("rejects an empty or garbage token", func(ctx SpecContext) {
			_, err := refresh(ctx, "")
			Expect(err).To(HaveConnectCode(connect.CodeInvalidArgument))

			_, err = refresh(ctx, "not.a.jwt")
			Expect(err).To(HaveConnectCode(connect.CodeUnauthenticated))
		})
	})

	Describe("Logout", func() {
		It("revokes the session of the given refresh token", func(ctx SpecContext) {
			loggedIn, err := login(ctx, user.Username, password)
			Expect(err).ShouldNot(HaveOccurred())
			jti := env.claims(ctx, loggedIn.GetToken()).ID

			_, err = handler.Logout(ctx, connect.NewRequest(&usersv1.LogoutRequest{Token: loggedIn.GetRefreshToken()}))
			Expect(err).ShouldNot(HaveOccurred())

			Expect(env.isRevoked(ctx, jti)).To(BeTrue())
		})

		It("revokes the session of the request's access token, and only that one", func(ctx SpecContext) {
			authCtx, claims := env.loginAs(ctx, user)
			_, other := env.loginAs(ctx, user)

			_, err := handler.Logout(authCtx, connect.NewRequest(&usersv1.LogoutRequest{}))
			Expect(err).ShouldNot(HaveOccurred())

			Expect(env.isRevoked(ctx, claims.ID)).To(BeTrue())
			Expect(env.isRevoked(ctx, other.ID)).To(BeFalse())
		})

		It("is unauthenticated without any token", func(ctx SpecContext) {
			_, err := handler.Logout(ctx, connect.NewRequest(&usersv1.LogoutRequest{}))
			Expect(err).To(HaveConnectCode(connect.CodeUnauthenticated))
		})
	})
})
