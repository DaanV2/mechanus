package campaigns_test

import (
	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/DaanV2/mechanus/server/pkg/database/models"
	campaign_service "github.com/DaanV2/mechanus/server/pkg/services/campaigns"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Campaign Service", func() {

	var (
		db      *database.DB
		service *campaign_service.Service
	)

	BeforeEach(func(setupCtx SpecContext) {
		db = util_test.CreateDatabase(setupCtx)
		service = campaign_service.NewService(db)
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
