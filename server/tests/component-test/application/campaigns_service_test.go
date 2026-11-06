package application_test

import (
	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/repositories"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Campaign Service", func() {

	var (
		db       *persistence.DB
		service  *application.CampaignService
		repo     *repositories.CampaignRepository
		userRepo *repositories.UserRepository
	)

	BeforeEach(func(setupCtx SpecContext) {
		db = util_test.CreateDatabase(setupCtx)
		repo = repositories.NewCampaignRepository(db)
		userRepo = repositories.NewUserRepository(db)
		service = application.NewCampaignService(repo)
	})

	Context("Get", func() {
		It("can get a campaign by id", func(ctx SpecContext) {
			campaign := &models.Campaign{
				Name: "Fellowship",
			}

			By("creating a new campaign")
			Expect(service.Create(ctx, campaign)).To(Succeed())
			Expect(campaign.ID).ToNot(BeEmpty())

			By("trying to retrieve the campaign")
			c, err := service.Get(ctx, campaign.ID)
			Expect(err).ToNot(HaveOccurred())

			Expect(c.ID).To(Equal(campaign.ID))
			Expect(c.Name).To(Equal(campaign.Name))
		})

		It("returns error if campaign does not exist", func(ctx SpecContext) {
			_, err := service.Get(ctx, "non-existent-id")
			Expect(err).To(HaveOccurred())
		})

		It("includes the users of the campaign", func(ctx SpecContext) {
			user := createUser(ctx, userRepo)
			campaign := &models.Campaign{Name: "Fellowship"}
			Expect(service.CreateWithUser(ctx, campaign, user)).To(Succeed())

			c, err := service.Get(ctx, campaign.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.Users).To(HaveLen(1))
			Expect(c.Users[0].ID).To(Equal(user.ID))
			Expect(c.Users[0].Username).To(Equal(user.Username))
		})
	})

	Context("CreateWithUser", func() {
		It("creates the campaign and links the user, without altering the user", func(ctx SpecContext) {
			user := createUser(ctx, userRepo)
			userID := user.ID

			campaign := &models.Campaign{Name: "Fellowship"}
			Expect(service.CreateWithUser(ctx, campaign, user)).To(Succeed())

			Expect(campaign.ID).ToNot(BeEmpty())
			Expect(user.ID).To(Equal(userID))
			Expect(campaign.Users).To(ContainElement(user))
			Expect(user.Campaigns).To(ContainElement(campaign))

			var count int64
			Expect(db.WithContext(ctx).Model(&models.User{}).Where("username = ?", user.Username).Count(&count).Error).To(Succeed())
			Expect(count).To(BeEquivalentTo(1), "linking should not create a duplicate user")
		})
	})

	Context("List", func() {
		It("returns all campaigns", func(ctx SpecContext) {
			userA, userB := createUser(ctx, userRepo), createUser(ctx, userRepo)
			campaignA := &models.Campaign{Name: "A"}
			campaignB := &models.Campaign{Name: "B"}
			Expect(service.CreateWithUser(ctx, campaignA, userA)).To(Succeed())
			Expect(service.CreateWithUser(ctx, campaignB, userB)).To(Succeed())

			campaigns, err := service.List(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(campaignIDs(campaigns)).To(ConsistOf(campaignA.ID, campaignB.ID))
		})
	})

	Context("ListForUser", func() {
		It("only returns the campaigns the user is in, including all their members", func(ctx SpecContext) {
			userA, userB := createUser(ctx, userRepo), createUser(ctx, userRepo)
			campaignA := &models.Campaign{Name: "A"}
			campaignB := &models.Campaign{Name: "B"}
			shared := &models.Campaign{Name: "Shared"}
			Expect(service.CreateWithUser(ctx, campaignA, userA)).To(Succeed())
			Expect(service.CreateWithUser(ctx, campaignB, userB)).To(Succeed())
			Expect(service.CreateWithUser(ctx, shared, userA)).To(Succeed())
			Expect(service.AddUsers(ctx, shared, userB)).To(Succeed())

			campaigns, err := service.ListForUser(ctx, userA.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(campaignIDs(campaigns)).To(ConsistOf(campaignA.ID, shared.ID))

			for _, c := range campaigns {
				if c.ID == shared.ID {
					Expect(c.Users).To(HaveLen(2), "all members should be loaded, not just the filtered user")
				}
			}
		})

		It("returns nothing for a user without campaigns", func(ctx SpecContext) {
			user := createUser(ctx, userRepo)

			campaigns, err := service.ListForUser(ctx, user.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(campaigns).To(BeEmpty())
		})
	})

	Context("CreateWithUser owner", func() {
		It("makes the user the owner of the campaign", func(ctx SpecContext) {
			user := createUser(ctx, userRepo)
			campaign := &models.Campaign{Name: "Fellowship"}
			Expect(service.CreateWithUser(ctx, campaign, user)).To(Succeed())

			c, err := service.Get(ctx, campaign.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.OwnerID).To(Equal(user.ID))
		})
	})

	Context("Update", func() {
		It("saves the name and description, also when the description is cleared", func(ctx SpecContext) {
			user := createUser(ctx, userRepo)
			campaign := &models.Campaign{Name: "Fellowship", Description: "One ring"}
			Expect(service.CreateWithUser(ctx, campaign, user)).To(Succeed())

			By("changing the name")
			campaign.Name = "Two Towers"
			Expect(service.Update(ctx, campaign)).To(Succeed())
			c, err := service.Get(ctx, campaign.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.Name).To(Equal("Two Towers"))
			Expect(c.Description).To(Equal("One ring"))

			By("clearing the description")
			campaign.Description = ""
			Expect(service.Update(ctx, campaign)).To(Succeed())
			c, err = service.Get(ctx, campaign.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.Description).To(BeEmpty())

			By("keeping the owner and players")
			Expect(c.OwnerID).To(Equal(user.ID))
			Expect(c.Users).To(HaveLen(1))
		})

		It("returns not found for unknown campaigns", func(ctx SpecContext) {
			err := service.Update(ctx, &models.Campaign{Model: models.Model{ID: "does-not-exist"}, Name: "x"})
			Expect(persistence.IsNotExist(err)).To(BeTrue())
		})
	})

	Context("Delete", func() {
		It("soft deletes the campaign", func(ctx SpecContext) {
			user := createUser(ctx, userRepo)
			campaign := &models.Campaign{Name: "Fellowship"}
			Expect(service.CreateWithUser(ctx, campaign, user)).To(Succeed())

			Expect(service.Delete(ctx, campaign.ID)).To(Succeed())

			By("no longer finding the campaign")
			_, err := service.Get(ctx, campaign.ID)
			Expect(persistence.IsNotExist(err)).To(BeTrue())

			campaigns, err := service.List(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(campaignIDs(campaigns)).ToNot(ContainElement(campaign.ID))

			campaigns, err = service.ListForUser(ctx, user.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(campaigns).To(BeEmpty())

			By("keeping the row in the database")
			var deleted models.Campaign
			Expect(db.WithContext(ctx).Unscoped().First(&deleted, "id = ?", campaign.ID).Error).To(Succeed())
			Expect(deleted.DeletedAt.Valid).To(BeTrue())
		})

		It("returns not found for unknown or already deleted campaigns", func(ctx SpecContext) {
			Expect(persistence.IsNotExist(service.Delete(ctx, "does-not-exist"))).To(BeTrue())

			campaign := &models.Campaign{Name: "Fellowship"}
			Expect(service.Create(ctx, campaign)).To(Succeed())
			Expect(service.Delete(ctx, campaign.ID)).To(Succeed())
			Expect(persistence.IsNotExist(service.Delete(ctx, campaign.ID))).To(BeTrue())
		})
	})

	Context("RemoveUser", func() {
		It("removes a player from the campaign", func(ctx SpecContext) {
			owner, player := createUser(ctx, userRepo), createUser(ctx, userRepo)
			campaign := &models.Campaign{Name: "Fellowship"}
			Expect(service.CreateWithUser(ctx, campaign, owner)).To(Succeed())
			Expect(service.AddUsers(ctx, campaign, player)).To(Succeed())

			Expect(service.RemoveUser(ctx, campaign, player)).To(Succeed())

			c, err := service.Get(ctx, campaign.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.Users).To(HaveLen(1))
			Expect(c.Users[0].ID).To(Equal(owner.ID))
		})

		It("refuses to remove the owner", func(ctx SpecContext) {
			owner := createUser(ctx, userRepo)
			campaign := &models.Campaign{Name: "Fellowship"}
			Expect(service.CreateWithUser(ctx, campaign, owner)).To(Succeed())

			Expect(service.RemoveUser(ctx, campaign, owner)).To(MatchError(application.ErrCannotRemoveOwner))

			c, err := service.Get(ctx, campaign.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.Users).To(HaveLen(1))
		})
	})
})

func createUser(ctx SpecContext, repo *repositories.UserRepository) *models.User {
	user := util_test.CreateUser()
	Expect(repo.Create(ctx, user)).To(Succeed())

	return user
}

func campaignIDs(campaigns []*models.Campaign) []string {
	ids := make([]string, 0, len(campaigns))
	for _, c := range campaigns {
		ids = append(ids, c.ID)
	}

	return ids
}
