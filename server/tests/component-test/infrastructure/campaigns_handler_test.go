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

	addPlayer := func(ctx context.Context, campaignId, username string) *campaignsv1.CampaignMember {
		GinkgoHelper()
		resp, err := handler.AddPlayer(ctx, connect.NewRequest(&campaignsv1.AddPlayerRequest{CampaignId: campaignId, Username: username}))
		Expect(err).ToNot(HaveOccurred())

		return resp.Msg.GetPlayer()
	}

	list := func(ctx context.Context) []string {
		GinkgoHelper()
		resp, err := handler.List(ctx, connect.NewRequest(&campaignsv1.ListCampaignsRequest{}))
		Expect(err).ToNot(HaveOccurred())

		ids := make([]string, 0, len(resp.Msg.GetCampaigns()))
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

		It("tells who can manage the campaign", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")
			Expect(campaign.GetOwnerId()).To(Equal(alice.ID))
			Expect(campaign.GetCanManage()).To(BeTrue())
			addPlayer(aliceContext(ctx), campaign.GetId(), bob.Username)

			canManage := func(userCtx context.Context) bool {
				resp, err := handler.Get(userCtx, connect.NewRequest(&campaignsv1.GetCampaignRequest{Id: campaign.GetId()}))
				Expect(err).ToNot(HaveOccurred())

				return resp.Msg.GetCampaign().GetCanManage()
			}
			Expect(canManage(aliceContext(ctx))).To(BeTrue(), "the owner can manage")
			Expect(canManage(adminContext(ctx))).To(BeTrue(), "admins can manage")
			Expect(canManage(bobContext(ctx))).To(BeFalse(), "players can't manage")
		})
	})

	Context("Update", func() {
		update := func(ctx context.Context, id, name, description string) (*campaignsv1.Campaign, error) {
			resp, err := handler.Update(ctx, connect.NewRequest(&campaignsv1.UpdateCampaignRequest{Id: id, Name: name, Description: description}))
			if err != nil {
				return nil, err
			}

			return resp.Msg.GetCampaign(), nil
		}

		It("lets the owner and admins change the name and description", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")

			updated, err := update(aliceContext(ctx), campaign.GetId(), " Two Towers ", "The journey continues")
			Expect(err).ToNot(HaveOccurred())
			Expect(updated.GetName()).To(Equal("Two Towers"))
			Expect(updated.GetDescription()).To(Equal("The journey continues"))

			updated, err = update(adminContext(ctx), campaign.GetId(), "Return of the King", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(updated.GetName()).To(Equal("Return of the King"))
			Expect(updated.GetDescription()).To(BeEmpty())
		})

		It("refuses players that aren't the owner", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")
			addPlayer(aliceContext(ctx), campaign.GetId(), bob.Username)

			_, err := update(bobContext(ctx), campaign.GetId(), "Bob's now", "")
			Expect(connect.CodeOf(err)).To(Equal(connect.CodePermissionDenied))
		})

		It("hides the campaign from non-members", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")

			_, err := update(bobContext(ctx), campaign.GetId(), "Bob's now", "")
			Expect(connect.CodeOf(err)).To(Equal(connect.CodeNotFound))
		})

		It("rejects an empty name", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")

			_, err := update(aliceContext(ctx), campaign.GetId(), " ", "")
			Expect(connect.CodeOf(err)).To(Equal(connect.CodeInvalidArgument))
		})

		It("rejects unauthenticated requests", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")

			_, err := update(ctx, campaign.GetId(), "Nope", "")
			Expect(connect.CodeOf(err)).To(Equal(connect.CodeUnauthenticated))
		})
	})

	Context("Delete", func() {
		remove := func(ctx context.Context, id string) error {
			_, err := handler.Delete(ctx, connect.NewRequest(&campaignsv1.DeleteCampaignRequest{Id: id}))

			return err
		}

		It("lets the owner delete the campaign", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")

			Expect(remove(aliceContext(ctx), campaign.GetId())).To(Succeed())

			_, err := handler.Get(aliceContext(ctx), connect.NewRequest(&campaignsv1.GetCampaignRequest{Id: campaign.GetId()}))
			Expect(connect.CodeOf(err)).To(Equal(connect.CodeNotFound))
			Expect(list(aliceContext(ctx))).To(BeEmpty())
			Expect(list(adminContext(ctx))).To(BeEmpty())
		})

		It("lets admins delete any campaign", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")

			Expect(remove(adminContext(ctx), campaign.GetId())).To(Succeed())
			Expect(list(aliceContext(ctx))).To(BeEmpty())
		})

		It("refuses players that aren't the owner, and hides it from non-members", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")
			other := create(aliceContext(ctx), "Other")
			addPlayer(aliceContext(ctx), campaign.GetId(), bob.Username)

			Expect(connect.CodeOf(remove(bobContext(ctx), campaign.GetId()))).To(Equal(connect.CodePermissionDenied))
			Expect(connect.CodeOf(remove(bobContext(ctx), other.GetId()))).To(Equal(connect.CodeNotFound))
			Expect(list(aliceContext(ctx))).To(ConsistOf(campaign.GetId(), other.GetId()))
		})

		It("rejects unauthenticated requests", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")

			Expect(connect.CodeOf(remove(ctx, campaign.GetId()))).To(Equal(connect.CodeUnauthenticated))
		})
	})

	Context("ListPlayers", func() {
		listPlayers := func(ctx context.Context, id string) ([]string, error) {
			resp, err := handler.ListPlayers(ctx, connect.NewRequest(&campaignsv1.ListPlayersRequest{CampaignId: id}))
			if err != nil {
				return nil, err
			}

			names := make([]string, 0, len(resp.Msg.GetPlayers()))
			for _, p := range resp.Msg.GetPlayers() {
				names = append(names, p.GetName())
			}

			return names, nil
		}

		It("shows the players to members and admins", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")
			addPlayer(aliceContext(ctx), campaign.GetId(), bob.Username)

			for _, userCtx := range []context.Context{aliceContext(ctx), bobContext(ctx), adminContext(ctx)} {
				Expect(listPlayers(userCtx, campaign.GetId())).To(ConsistOf(alice.Username, bob.Username))
			}
		})

		It("hides the players from non-members", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")

			_, err := listPlayers(bobContext(ctx), campaign.GetId())
			Expect(connect.CodeOf(err)).To(Equal(connect.CodeNotFound))
		})
	})

	Context("AddPlayer", func() {
		add := func(ctx context.Context, id, username string) error {
			_, err := handler.AddPlayer(ctx, connect.NewRequest(&campaignsv1.AddPlayerRequest{CampaignId: id, Username: username}))

			return err
		}

		It("lets the owner and admins add existing users", func(ctx SpecContext) {
			carol := util_test.CreateUser()
			Expect(userRepo.Create(ctx, carol)).To(Succeed())
			campaign := create(aliceContext(ctx), "Fellowship")

			player := addPlayer(aliceContext(ctx), campaign.GetId(), bob.Username)
			Expect(player.GetId()).To(Equal(bob.ID))
			Expect(add(adminContext(ctx), campaign.GetId(), carol.Username)).To(Succeed())

			By("the new players seeing the campaign")
			Expect(list(bobContext(ctx))).To(ConsistOf(campaign.GetId()))
			Expect(list(withUser(carol)(ctx))).To(ConsistOf(campaign.GetId()))
		})

		It("refuses players that aren't the owner", func(ctx SpecContext) {
			carol := util_test.CreateUser()
			Expect(userRepo.Create(ctx, carol)).To(Succeed())
			campaign := create(aliceContext(ctx), "Fellowship")
			addPlayer(aliceContext(ctx), campaign.GetId(), bob.Username)

			Expect(connect.CodeOf(add(bobContext(ctx), campaign.GetId(), carol.Username))).To(Equal(connect.CodePermissionDenied))
		})

		It("hides the campaign from non-members", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")

			Expect(connect.CodeOf(add(bobContext(ctx), campaign.GetId(), bob.Username))).To(Equal(connect.CodeNotFound))
			Expect(list(bobContext(ctx))).To(BeEmpty())
		})

		It("reports unknown users, existing players and empty usernames", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")

			Expect(connect.CodeOf(add(aliceContext(ctx), campaign.GetId(), "nobody-by-this-name"))).To(Equal(connect.CodeNotFound))
			Expect(connect.CodeOf(add(aliceContext(ctx), campaign.GetId(), alice.Username))).To(Equal(connect.CodeAlreadyExists))
			Expect(connect.CodeOf(add(aliceContext(ctx), campaign.GetId(), " "))).To(Equal(connect.CodeInvalidArgument))
		})
	})

	Context("RemovePlayer", func() {
		removePlayer := func(ctx context.Context, id, userId string) error {
			_, err := handler.RemovePlayer(ctx, connect.NewRequest(&campaignsv1.RemovePlayerRequest{CampaignId: id, UserId: userId}))

			return err
		}

		It("lets the owner and admins remove players", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")
			addPlayer(aliceContext(ctx), campaign.GetId(), bob.Username)
			Expect(removePlayer(aliceContext(ctx), campaign.GetId(), bob.ID)).To(Succeed())
			Expect(list(bobContext(ctx))).To(BeEmpty())

			addPlayer(aliceContext(ctx), campaign.GetId(), bob.Username)
			Expect(removePlayer(adminContext(ctx), campaign.GetId(), bob.ID)).To(Succeed())
			Expect(list(bobContext(ctx))).To(BeEmpty())
		})

		It("lets players leave, but not remove others", func(ctx SpecContext) {
			carol := util_test.CreateUser()
			Expect(userRepo.Create(ctx, carol)).To(Succeed())
			campaign := create(aliceContext(ctx), "Fellowship")
			addPlayer(aliceContext(ctx), campaign.GetId(), bob.Username)
			addPlayer(aliceContext(ctx), campaign.GetId(), carol.Username)

			Expect(connect.CodeOf(removePlayer(bobContext(ctx), campaign.GetId(), carol.ID))).To(Equal(connect.CodePermissionDenied))
			Expect(removePlayer(bobContext(ctx), campaign.GetId(), bob.ID)).To(Succeed())
			Expect(list(bobContext(ctx))).To(BeEmpty())
			Expect(list(withUser(carol)(ctx))).To(ConsistOf(campaign.GetId()))
		})

		It("never removes the owner", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")

			Expect(connect.CodeOf(removePlayer(aliceContext(ctx), campaign.GetId(), alice.ID))).To(Equal(connect.CodeFailedPrecondition))
			Expect(connect.CodeOf(removePlayer(adminContext(ctx), campaign.GetId(), alice.ID))).To(Equal(connect.CodeFailedPrecondition))
			Expect(list(aliceContext(ctx))).To(ConsistOf(campaign.GetId()))
		})

		It("reports users that aren't players, and hides the campaign from non-members", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")

			Expect(connect.CodeOf(removePlayer(aliceContext(ctx), campaign.GetId(), bob.ID))).To(Equal(connect.CodeNotFound))
			Expect(connect.CodeOf(removePlayer(bobContext(ctx), campaign.GetId(), bob.ID))).To(Equal(connect.CodeNotFound))
		})

		It("rejects unauthenticated requests", func(ctx SpecContext) {
			campaign := create(aliceContext(ctx), "Fellowship")

			Expect(connect.CodeOf(removePlayer(ctx, campaign.GetId(), alice.ID))).To(Equal(connect.CodeUnauthenticated))
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
