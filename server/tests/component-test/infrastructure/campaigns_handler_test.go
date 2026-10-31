package infrastructure_test

import (
	"context"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/repositories"
	"github.com/DaanV2/mechanus/server/infrastructure/transport/grpc"
	campaignsv1 "github.com/DaanV2/mechanus/server/proto/campaigns/v1"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CampaignServiceHandler", func() {
	var (
		handler      *grpc.CampaignServiceHandler
		userRepo     *repositories.UserRepository
		alice, bob   *models.User
		admin        *models.User
		aliceContext func(context.Context) context.Context
		bobContext   func(context.Context) context.Context
		adminContext func(context.Context) context.Context
	)

	BeforeEach(func(setupCtx SpecContext) {
		db := util_test.CreateDatabase(setupCtx)
		userRepo = repositories.NewUserRepository(db)
		handler = grpc.NewCampaignServiceHandler(
			application.NewCampaignService(repositories.NewCampaignRepository(db)),
			application.NewUserService(userRepo),
		)

		alice, bob, admin = util_test.CreateUser(), util_test.CreateUser(), util_test.CreateUser()
		admin.Roles = []string{roles.Admin.String()}
		for _, u := range []*models.User{alice, bob, admin} {
			Expect(userRepo.Create(setupCtx, u)).To(Succeed())
		}

		aliceContext = withUser(alice)
		bobContext = withUser(bob)
		adminContext = withUser(admin)
	})

	create := func(ctx context.Context, name string) *campaignsv1.Campaign {
		GinkgoHelper()
		resp, err := handler.Create(ctx, connect.NewRequest(&campaignsv1.CreateCampaignRequest{Name: name}))
		Expect(err).ToNot(HaveOccurred())

		return resp.Msg.GetCampaign()
	}

	list := func(ctx context.Context) []string {
		GinkgoHelper()
		resp, err := handler.List(ctx, connect.NewRequest(&campaignsv1.ListCampaignsRequest{}))
		Expect(err).ToNot(HaveOccurred())

		ids := []string{}
		for _, c := range resp.Msg.GetCampaigns() {
			ids = append(ids, c.GetId())
		}

		return ids
	}

	Context("Create", func() {
		It("makes the creator a member", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")

			Expect(campaign.GetId()).ToNot(BeEmpty())
			Expect(campaign.GetName()).To(Equal("Fellowship"))
			Expect(campaign.GetMembers()).To(HaveLen(1))
			Expect(campaign.GetMembers()[0].GetId()).To(Equal(alice.ID))
			Expect(campaign.GetMembers()[0].GetName()).To(Equal(alice.Username))
		})

		It("rejects an empty name", func(ctx SpecContext) {
			_, err := handler.Create(aliceContext(ctx), connect.NewRequest(&campaignsv1.CreateCampaignRequest{Name: "  "}))
			Expect(connect.CodeOf(err)).To(Equal(connect.CodeInvalidArgument))
		})

		It("rejects unauthenticated requests", func(ctx SpecContext) {
			_, err := handler.Create(ctx, connect.NewRequest(&campaignsv1.CreateCampaignRequest{Name: "Fellowship"}))
			Expect(connect.CodeOf(err)).To(Equal(connect.CodeUnauthenticated))
		})
	})

	Context("List", func() {
		It("only shows users the campaigns they are in, and admins all of them", func(ctx SpecContext) {
			aliceCampaign := create(aliceContext(ctx), "Alice's")
			bobCampaign := create(bobContext(ctx), "Bob's")

			Expect(list(aliceContext(ctx))).To(ConsistOf(aliceCampaign.GetId()))
			Expect(list(bobContext(ctx))).To(ConsistOf(bobCampaign.GetId()))
			Expect(list(adminContext(ctx))).To(ConsistOf(aliceCampaign.GetId(), bobCampaign.GetId()))
		})

		It("rejects unauthenticated requests", func(ctx SpecContext) {
			_, err := handler.List(ctx, connect.NewRequest(&campaignsv1.ListCampaignsRequest{}))
			Expect(connect.CodeOf(err)).To(Equal(connect.CodeUnauthenticated))
		})
	})

	Context("Get", func() {
		It("returns the campaign to its members and admins", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")

			for _, userCtx := range []context.Context{aliceContext(ctx), adminContext(ctx)} {
				resp, err := handler.Get(userCtx, connect.NewRequest(&campaignsv1.GetCampaignRequest{Id: campaign.GetId()}))
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Msg.GetCampaign().GetName()).To(Equal("Fellowship"))
				Expect(resp.Msg.GetCampaign().GetCreatedAt().AsTime()).ToNot(BeZero())
			}
		})

		It("hides the campaign from non-members", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")

			_, err := handler.Get(bobContext(ctx), connect.NewRequest(&campaignsv1.GetCampaignRequest{Id: campaign.GetId()}))
			Expect(connect.CodeOf(err)).To(Equal(connect.CodeNotFound))
		})

		It("returns not found for unknown campaigns", func(ctx SpecContext) {
			_, err := handler.Get(adminContext(ctx), connect.NewRequest(&campaignsv1.GetCampaignRequest{Id: "does-not-exist"}))
			Expect(connect.CodeOf(err)).To(Equal(connect.CodeNotFound))
		})
	})
})

func withUser(user *models.User) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return authentication.ContextWithJWT(ctx, &authentication.JWTClaims{
			User: authentication.JWTUser{
				ID:    user.ID,
				Name:  user.Username,
				Roles: user.Roles,
			},
		}, true)
	}
}
