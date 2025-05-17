package users_test

import (
	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/DaanV2/mechanus/server/pkg/database/models"
	user_service "github.com/DaanV2/mechanus/server/pkg/services/users"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("User Service", func() {

	var (
		db      *database.DB
		service *user_service.Service
	)

	BeforeEach(func() {
		db = util_test.CreateDatabase()
		service = user_service.NewService(db)
	})

	It("can create a new user", func(ctx SpecContext) {
		user := &models.User{
			Name:         "gandalf",
			Roles:        []string{"wizard"},
			PasswordHash: []byte("the-one-ring"),
		}

		By("creating a new user")
		Expect(service.Create(ctx, user)).To(Succeed())
		Expect(user.ID).ToNot(BeEmpty())

		By("trying to retrieve the user")
		u, err := service.Get(ctx, user.ID)
		Expect(err).ToNot(HaveOccurred())

		Expect(user.ID).To(Equal(u.ID))
		Expect(user.Name).To(Equal(u.Name))
		Expect(user.PasswordHash).To(Equal(u.PasswordHash))
		Expect(u.PasswordHash).ToNot(Equal([]byte("the-one-ring")))
		Expect(u.Roles).ToNot(Equal([]string{"wizard"}))
	})

	It("can get a user by name", func(ctx SpecContext) {
		user := &models.User{
			Name: "gandalf",
		}

		By("creating a new user")
		Expect(service.Create(ctx, user)).To(Succeed())
		Expect(user.ID).ToNot(BeEmpty())

		By("trying to retrieve the user")
		u, err := service.GetByUsername(ctx, "gandalf")
		Expect(err).ToNot(HaveOccurred())

		Expect(user.ID).To(Equal(u.ID))
		Expect(user.Name).To(Equal(u.Name))
	})

})
