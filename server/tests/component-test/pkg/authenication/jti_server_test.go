package authenication_test

import (
	"github.com/DaanV2/mechanus/server/pkg/authenication"
	"github.com/DaanV2/mechanus/server/pkg/database"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JtiServer", func() {
	var (
		db      *database.DB
		service *authenication.JTIService
	)

	BeforeEach(func() {
		db = util_test.CreateDatabase()
		service = authenication.NewJTIService(db)
	})

	Context("GetActiveOrCreate", func() {
		It("creates a new JTI if no exist", func(ctx SpecContext) {
			userId := util_test.CreateUserID()

			jti, err := service.GetActiveOrCreate(ctx, userId)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(jti).ToNot(BeNil())
			Expect(jti.ID).ToNot(BeEmpty())
			Expect(jti.UserID).To(Equal(userId))
			Expect(jti.Revoked).To(BeFalse())
			Expect(jti.Valid()).To(BeTrue())
		})

		It("returns the existing JTI if it is still valid", func(ctx SpecContext) {
			userId := util_test.CreateUserID()
			// Create the first JTI
			firstJTI, err := service.GetActiveOrCreate(ctx, userId)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(firstJTI).ToNot(BeNil())

			// Call again, should return the same JTI
			secondJTI, err := service.GetActiveOrCreate(ctx, userId)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(secondJTI).ToNot(BeNil())
			Expect(secondJTI.ID).To(Equal(firstJTI.ID))
		})

		It("creates a new JTI if the existing one is revoked", func(ctx SpecContext) {
			userId := util_test.CreateUserID()
			// Create the first JTI
			firstJTI, err := service.GetActiveOrCreate(ctx, userId)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(firstJTI).ToNot(BeNil())

			// Revoke the first JTI
			_, err = service.Revoke(ctx, firstJTI.ID)
			Expect(err).ShouldNot(HaveOccurred())

			// Call again, should create a new JTI
			secondJTI, err := service.GetActiveOrCreate(ctx, userId)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(secondJTI).ToNot(BeNil())
			Expect(secondJTI.ID).ToNot(Equal(firstJTI.ID))
		})
	})

	Context("GetByUser", func() {
		It("returns all JTIs for a user", func(ctx SpecContext) {
			userId := util_test.CreateUserID()
			_, err := service.Create(ctx, userId)
			Expect(err).ShouldNot(HaveOccurred())
			_, err = service.Create(ctx, userId)
			Expect(err).ShouldNot(HaveOccurred())

			jtis, err := service.GetByUser(ctx, userId)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(jtis).To(HaveLen(2))
			for _, jti := range jtis {
				Expect(jti.UserID).To(Equal(userId))
			}
		})

		It("returns error if userId is empty", func(ctx SpecContext) {
			jtis, err := service.GetByUser(ctx, "")
			Expect(err).To(HaveOccurred())
			Expect(jtis).To(BeNil())
		})
	})

	Context("GetActive", func() {
		It("returns only active (not revoked) JTIs", func(ctx SpecContext) {
			userId := util_test.CreateUserID()
			active, err := service.Create(ctx, userId)
			Expect(err).ShouldNot(HaveOccurred())
			_, err = service.Revoke(ctx, active.ID)
			Expect(err).ShouldNot(HaveOccurred())
			// Create another active JTI
			active2, err := service.Create(ctx, userId)
			Expect(err).ShouldNot(HaveOccurred())

			jtis, err := service.GetActive(ctx, userId)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(jtis).To(HaveLen(1))
			Expect(jtis[0].ID).To(Equal(active2.ID))
		})

		It("does not return revoked JTIs", func(ctx SpecContext) {
			userId := util_test.CreateUserID()
			jti1, err := service.Create(ctx, userId)
			Expect(err).ShouldNot(HaveOccurred())
			jti2, err := service.Create(ctx, userId)
			Expect(err).ShouldNot(HaveOccurred())
			// Revoke both JTIs
			_, err = service.Revoke(ctx, jti1.ID)
			Expect(err).ShouldNot(HaveOccurred())
			_, err = service.Revoke(ctx, jti2.ID)
			Expect(err).ShouldNot(HaveOccurred())

			jtis, err := service.GetActive(ctx, userId)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(jtis).To(BeEmpty())
		})

		It("returns error if userId is empty", func(ctx SpecContext) {
			jtis, err := service.GetActive(ctx, "")
			Expect(err).To(HaveOccurred())
			Expect(jtis).To(BeNil())
		})
	})

	Context("Get", func() {
		It("returns the JTI by ID", func(ctx SpecContext) {
			userId := util_test.CreateUserID()
			created, err := service.Create(ctx, userId)
			Expect(err).ShouldNot(HaveOccurred())

			fetched, err := service.Get(ctx, created.ID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(fetched).ToNot(BeNil())
			Expect(fetched.ID).To(Equal(created.ID))
		})

		It("returns error if jti is empty", func(ctx SpecContext) {
			jti, err := service.Get(ctx, "")
			Expect(err).To(HaveOccurred())
			Expect(jti).To(BeNil())
		})
	})

	Context("Create", func() {
		It("creates a new JTI for a user", func(ctx SpecContext) {
			userId := util_test.CreateUserID()
			jti, err := service.Create(ctx, userId)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(jti).ToNot(BeNil())
			Expect(jti.UserID).To(Equal(userId))
			Expect(jti.Revoked).To(BeFalse())
		})

		It("returns error if userId is empty", func(ctx SpecContext) {
			jti, err := service.Create(ctx, "")
			Expect(err).To(HaveOccurred())
			Expect(jti).To(BeNil())
		})
	})

	Context("Revoke", func() {
		It("revokes a JTI by ID", func(ctx SpecContext) {
			userId := util_test.CreateUserID()
			jti, err := service.Create(ctx, userId)
			Expect(err).ShouldNot(HaveOccurred())

			revoked, err := service.Revoke(ctx, jti.ID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(revoked).To(BeTrue())
		})

		It("returns error if jti is empty", func(ctx SpecContext) {
			revoked, err := service.Revoke(ctx, "")
			Expect(err).To(HaveOccurred())
			Expect(revoked).To(BeFalse())
		})
	})

})
