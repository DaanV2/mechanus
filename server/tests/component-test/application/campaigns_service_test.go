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
		db      *persistence.DB
		service *application.CampaignService
		repo    *repositories.CampaignRepository
	)

	BeforeEach(func(setupCtx SpecContext) {
		db = util_test.CreateDatabase(setupCtx)
		repo = repositories.NewCampaignRepository(db)
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
	})
})
