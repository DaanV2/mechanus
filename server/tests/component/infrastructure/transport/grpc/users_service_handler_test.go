package grpc_test

import (
	"context"

	"connectrpc.com/connect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/grpc"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xcrypto"
	usersv1 "github.com/DaanV2/mechanus/server/proto/users/v1"
)

var _ = Describe("UserServiceHandler", func() {
	const password = "hunter2-but-longer"

	var (
		env     *authEnv
		handler *grpc.UserServiceHandler

		admin, user, other                   *models.User
		adminClaims, userClaims, otherClaims *authentication.JWTClaims
	)

	BeforeEach(func(ctx SpecContext) {
		env = newAuthEnv(ctx)
		handler = grpc.NewUserServiceHandler(env.users, env.jtis, grpc.UserServiceOptions{})

		admin = env.createUser(ctx, password, "admin")
		user = env.createUser(ctx, password, "user")
		other = env.createUser(ctx, password, "user")
		_, adminClaims = env.loginAs(ctx, admin)
		_, userClaims = env.loginAs(ctx, user)
		_, otherClaims = env.loginAs(ctx, other)
	})

	// as returns a context authenticated as the session of the claims
	as := func(ctx context.Context, claims *authentication.JWTClaims) context.Context {
		return authentication.ContextWithJWT(ctx, claims, true)
	}

	Describe("Create", func() {
		create := func(ctx context.Context) error {
			_, err := handler.Create(ctx, connect.NewRequest(&usersv1.CreateAccountRequest{Username: "new_" + user.Username, Password: password}))

			return err
		}

		It("requires an admin while sign-up is disabled", func(ctx SpecContext) {
			Expect(create(ctx)).To(HaveConnectCode(connect.CodeUnauthenticated))
			Expect(create(as(ctx, userClaims))).To(HaveConnectCode(connect.CodePermissionDenied))
			Expect(create(as(ctx, adminClaims))).To(Succeed())
		})

		It("allows anyone while sign-up is enabled", func(ctx SpecContext) {
			handler = grpc.NewUserServiceHandler(env.users, env.jtis, grpc.UserServiceOptions{SignupEnabled: true})

			Expect(create(ctx)).To(Succeed())
		})
	})

	Describe("Get", func() {
		get := func(ctx context.Context, id string) (*usersv1.User, error) {
			resp, err := handler.Get(ctx, connect.NewRequest(&usersv1.GetUserRequest{Id: id}))
			if err != nil {
				return nil, err
			}

			return resp.Msg.GetUser(), nil
		}

		It("is unauthenticated for anonymous requests", func(ctx SpecContext) {
			_, err := get(ctx, user.ID)
			Expect(err).To(HaveConnectCode(connect.CodeUnauthenticated))
		})

		It("is unauthenticated for a request with an invalid token", func(ctx SpecContext) {
			_, err := get(authentication.ContextWithJWT(ctx, userClaims, false), user.ID)
			Expect(err).To(HaveConnectCode(connect.CodeUnauthenticated))
		})

		It("returns full info to the user themselves and admins", func(ctx SpecContext) {
			for _, authCtx := range []context.Context{as(ctx, userClaims), as(ctx, adminClaims)} {
				result, err := get(authCtx, user.ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(result.GetRoles()).To(ConsistOf("user"))
			}
		})

		It("returns reduced info to other users", func(ctx SpecContext) {
			result, err := get(as(ctx, otherClaims), user.ID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result.GetName()).To(Equal(user.Username))
			Expect(result.GetRoles()).To(BeEmpty())
		})
	})

	Describe("List", func() {
		It("requires an admin", func(ctx SpecContext) {
			_, err := handler.List(ctx, connect.NewRequest(&usersv1.ListUsersRequest{}))
			Expect(err).To(HaveConnectCode(connect.CodeUnauthenticated))

			_, err = handler.List(as(ctx, userClaims), connect.NewRequest(&usersv1.ListUsersRequest{}))
			Expect(err).To(HaveConnectCode(connect.CodePermissionDenied))

			resp, err := handler.List(as(ctx, adminClaims), connect.NewRequest(&usersv1.ListUsersRequest{}))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Msg.GetUsers()).To(HaveLen(3))
		})
	})

	Describe("ResetPassword", func() {
		reset := func(ctx context.Context) error {
			_, err := handler.ResetPassword(ctx, connect.NewRequest(&usersv1.ResetPasswordRequest{Id: user.ID, NewPassword: "new password"}))

			return err
		}

		It("requires an admin", func(ctx SpecContext) {
			Expect(reset(as(ctx, userClaims))).To(HaveConnectCode(connect.CodePermissionDenied))
		})

		It("signs out every session of the user", func(ctx SpecContext) {
			Expect(reset(as(ctx, adminClaims))).To(Succeed())

			Expect(env.isRevoked(ctx, userClaims.ID)).To(BeTrue())
		})
	})

	Describe("ChangePassword", func() {
		change := func(ctx context.Context, current string) error {
			_, err := handler.ChangePassword(ctx, connect.NewRequest(&usersv1.ChangePasswordRequest{CurrentPassword: current, NewPassword: "a new password"}))

			return err
		}

		It("is unauthenticated for anonymous requests", func(ctx SpecContext) {
			Expect(change(ctx, password)).To(HaveConnectCode(connect.CodeUnauthenticated))
		})

		It("rejects a wrong current password", func(ctx SpecContext) {
			Expect(change(as(ctx, userClaims), "wrong")).To(HaveConnectCode(connect.CodePermissionDenied))
		})

		It("changes the password and signs out the other sessions", func(ctx SpecContext) {
			_, otherSession := env.loginAs(ctx, user)

			Expect(change(as(ctx, userClaims), password)).To(Succeed())

			Expect(env.isRevoked(ctx, userClaims.ID)).To(BeFalse())
			Expect(env.isRevoked(ctx, otherSession.ID)).To(BeTrue())

			stored, err := env.users.Get(ctx, user.ID)
			Expect(err).ShouldNot(HaveOccurred())
			ok, err := xcrypto.ComparePassword(stored.PasswordHash, []byte("a new password"))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ok).To(BeTrue())
		})
	})

	Describe("SetRoles", func() {
		setRoles := func(ctx context.Context, id string, roles ...string) (*usersv1.User, error) {
			resp, err := handler.SetRoles(ctx, connect.NewRequest(&usersv1.SetRolesRequest{Id: id, Roles: roles}))
			if err != nil {
				return nil, err
			}

			return resp.Msg.GetUser(), nil
		}

		It("requires an admin", func(ctx SpecContext) {
			_, err := setRoles(as(ctx, userClaims), user.ID, "admin")
			Expect(err).To(HaveConnectCode(connect.CodePermissionDenied))
		})

		It("replaces the roles of the user", func(ctx SpecContext) {
			result, err := setRoles(as(ctx, adminClaims), user.ID, "operator", "Viewer", "operator")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result.GetRoles()).To(ConsistOf("operator", "viewer"))

			stored, err := env.users.Get(ctx, user.ID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect([]string(stored.Roles)).To(ConsistOf("operator", "viewer"))
		})

		DescribeTable("rejects",
			func(ctx SpecContext, roles []string, self bool, code connect.Code) {
				id := user.ID
				if self {
					id = admin.ID
				}

				_, err := setRoles(as(ctx, adminClaims), id, roles...)
				Expect(err).To(HaveConnectCode(code))
			},
			Entry("unknown roles", []string{"wizard"}, false, connect.CodeInvalidArgument),
			Entry("no roles", []string{}, false, connect.CodeInvalidArgument),
			Entry("removing your own admin role", []string{"user"}, true, connect.CodeFailedPrecondition),
		)

		It("returns not found for an unknown user", func(ctx SpecContext) {
			_, err := setRoles(as(ctx, adminClaims), "does-not-exist", "user")
			Expect(err).To(HaveConnectCode(connect.CodeNotFound))
		})
	})
})
