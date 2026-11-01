package grpc_test

import (
	"context"
	"sync"
	"time"

	"connectrpc.com/connect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/repositories"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/grpc"
	usersv1 "github.com/DaanV2/mechanus/server/proto/users/v1"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
)

var _ = Describe("InviteServiceHandler", func() {
	const password = "hunter2-but-longer"

	var (
		env     *authEnv
		repo    *repositories.InviteRepository
		handler *grpc.InviteServiceHandler

		adminClaims, userClaims *authentication.JWTClaims
	)

	BeforeEach(func(ctx SpecContext) {
		env = newAuthEnv(ctx)
		repo = repositories.NewInviteRepository(env.db)
		handler = grpc.NewInviteServiceHandler(application.NewInviteService(repo, env.users, time.Hour))

		_, adminClaims = env.loginAs(ctx, env.createUser(ctx, password, "admin"))
		_, userClaims = env.loginAs(ctx, env.createUser(ctx, password, "user"))
	})

	as := func(claims *authentication.JWTClaims) context.Context {
		return authentication.ContextWithJWT(context.Background(), claims, true)
	}

	create := func(ctx context.Context, validFor time.Duration) (*usersv1.Invite, error) {
		resp, err := handler.Create(ctx, connect.NewRequest(&usersv1.CreateInviteRequest{ValidForSeconds: int64(validFor.Seconds())}))
		if err != nil {
			return nil, err
		}

		return resp.Msg.GetInvite(), nil
	}

	accept := func(ctx context.Context, id, username string) (*usersv1.User, error) {
		resp, err := handler.Accept(ctx, connect.NewRequest(&usersv1.AcceptInviteRequest{Id: id, Username: username, Password: password}))
		if err != nil {
			return nil, err
		}

		return resp.Msg.GetUser(), nil
	}

	check := func(ctx context.Context, id string) bool {
		resp, err := handler.Check(ctx, connect.NewRequest(&usersv1.CheckInviteRequest{Id: id}))
		Expect(err).ShouldNot(HaveOccurred())

		return resp.Msg.GetValid()
	}

	Describe("Create", func() {
		It("requires an admin", func(ctx SpecContext) {
			_, err := create(ctx, 0)
			Expect(err).To(HaveConnectCode(connect.CodeUnauthenticated))
			_, err = create(as(userClaims), 0)
			Expect(err).To(HaveConnectCode(connect.CodePermissionDenied))
		})

		It("uses the default duration, and caps long ones", func() {
			invite, err := create(as(adminClaims), 0)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(invite.GetExpiresAt().AsTime()).To(BeTemporally("~", time.Now().Add(time.Hour), time.Minute))
			Expect(invite.GetCreatedBy()).To(Equal(adminClaims.User.ID))

			invite, err = create(as(adminClaims), 365*24*time.Hour)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(invite.GetExpiresAt().AsTime()).To(BeTemporally("~", time.Now().Add(application.MaxInviteDuration), time.Minute))
		})
	})

	Describe("List and Revoke", func() {
		It("require an admin", func(ctx SpecContext) {
			_, err := handler.List(as(userClaims), connect.NewRequest(&usersv1.ListInvitesRequest{}))
			Expect(err).To(HaveConnectCode(connect.CodePermissionDenied))
			_, err = handler.Revoke(ctx, connect.NewRequest(&usersv1.RevokeInviteRequest{Id: "x"}))
			Expect(err).To(HaveConnectCode(connect.CodeUnauthenticated))
		})

		It("lists only usable invites, and revoked invites can't be used", func(ctx SpecContext) {
			kept, err := create(as(adminClaims), 0)
			Expect(err).ShouldNot(HaveOccurred())
			revoked, err := create(as(adminClaims), 0)
			Expect(err).ShouldNot(HaveOccurred())
			used, err := create(as(adminClaims), 0)
			Expect(err).ShouldNot(HaveOccurred())
			_, err = accept(ctx, used.GetId(), "user_"+util_test.CreateUserID())
			Expect(err).ShouldNot(HaveOccurred())

			_, err = handler.Revoke(as(adminClaims), connect.NewRequest(&usersv1.RevokeInviteRequest{Id: revoked.GetId()}))
			Expect(err).ShouldNot(HaveOccurred())
			_, err = handler.Revoke(as(adminClaims), connect.NewRequest(&usersv1.RevokeInviteRequest{Id: revoked.GetId()}))
			Expect(err).To(HaveConnectCode(connect.CodeNotFound))

			resp, err := handler.List(as(adminClaims), connect.NewRequest(&usersv1.ListInvitesRequest{}))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Msg.GetInvites()).To(HaveLen(1))
			Expect(resp.Msg.GetInvites()[0].GetId()).To(Equal(kept.GetId()))

			Expect(check(ctx, revoked.GetId())).To(BeFalse())
			_, err = accept(ctx, revoked.GetId(), "user_"+util_test.CreateUserID())
			Expect(err).To(HaveConnectCode(connect.CodeFailedPrecondition))
		})
	})

	Describe("Accept", func() {
		It("creates a user account without login, once", func(ctx SpecContext) {
			invite, err := create(as(adminClaims), 0)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(check(ctx, invite.GetId())).To(BeTrue())

			username := "user_" + util_test.CreateUserID()
			user, err := accept(ctx, invite.GetId(), username)
			Expect(err).ShouldNot(HaveOccurred())

			stored, err := env.users.Get(ctx, user.GetId())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(stored.Username).To(Equal(username))
			Expect([]string(stored.Roles)).To(Equal([]string{"user"}))

			saved, err := repo.Get(ctx, invite.GetId())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(saved.UsedBy).To(Equal(user.GetId()))

			Expect(check(ctx, invite.GetId())).To(BeFalse())
			_, err = accept(ctx, invite.GetId(), "user_"+util_test.CreateUserID())
			Expect(err).To(HaveConnectCode(connect.CodeFailedPrecondition))
		})

		It("rejects unknown and expired invites", func(ctx SpecContext) {
			_, err := accept(ctx, "does-not-exist", "user_"+util_test.CreateUserID())
			Expect(err).To(HaveConnectCode(connect.CodeFailedPrecondition))

			expired := &models.Invite{CreatedBy: adminClaims.User.ID, ExpiresAt: time.Now().Add(-time.Minute)}
			Expect(repo.Create(ctx, expired)).To(Succeed())
			Expect(check(ctx, expired.ID)).To(BeFalse())
			_, err = accept(ctx, expired.ID, "user_"+util_test.CreateUserID())
			Expect(err).To(HaveConnectCode(connect.CodeFailedPrecondition))
		})

		It("keeps the invite usable when the username is taken", func(ctx SpecContext) {
			invite, err := create(as(adminClaims), 0)
			Expect(err).ShouldNot(HaveOccurred())
			taken := env.createUser(ctx, password, "user")

			_, err = accept(ctx, invite.GetId(), taken.Username)
			Expect(err).To(HaveConnectCode(connect.CodeAlreadyExists))
			Expect(check(ctx, invite.GetId())).To(BeTrue())
		})

		It("requires a username and password", func(ctx SpecContext) {
			invite, err := create(as(adminClaims), 0)
			Expect(err).ShouldNot(HaveOccurred())

			_, err = accept(ctx, invite.GetId(), "")
			Expect(err).To(HaveConnectCode(connect.CodeInvalidArgument))
			Expect(check(ctx, invite.GetId())).To(BeTrue())
		})

		It("lets only one of concurrent accepts use the invite", func(ctx SpecContext) {
			invite, err := create(as(adminClaims), 0)
			Expect(err).ShouldNot(HaveOccurred())

			const attempts = 5
			var (
				wg        sync.WaitGroup
				mu        sync.Mutex
				succeeded int
			)
			for range attempts {
				wg.Go(func() {
					if _, err := accept(ctx, invite.GetId(), "user_"+util_test.CreateUserID()); err == nil {
						mu.Lock()
						succeeded++
						mu.Unlock()
					}
				})
			}
			wg.Wait()

			Expect(succeeded).To(Equal(1))
		})
	})
})
