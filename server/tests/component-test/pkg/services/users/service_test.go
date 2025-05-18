package users_test

import (
	"time"

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

	Context("create", func() {
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
			Expect(u.Roles).To(BeEquivalentTo([]string{"wizard"}))
		})
	})

	Context("Get", func() {
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

	Context("Update", func() {
		It("can update the user, but will skip the password", func(ctx SpecContext) {
			user := &models.User{
				Name:         "first",
				Roles:        []string{"first"},
				PasswordHash: []byte("first"),
			}

			By("creating a new user")
			Expect(service.Create(ctx, user)).To(Succeed())
			Expect(user.ID).ToNot(BeEmpty())

			updateduser := &models.User{
				Model:        user.Model,
				Name:         "second",
				Roles:        []string{"second"},
				PasswordHash: []byte("second"),
			}

			// Wait a bit so the UpdateAt timestamp can be increased
			time.Sleep(100 * time.Millisecond)

			By("updating the existing user")
			Expect(service.Update(ctx, updateduser)).To(Succeed())

			By("checking the new records")
			check, err := service.Get(ctx, user.ID)
			Expect(err).ToNot(HaveOccurred())

			Expect(check.ID).To(Equal(user.ID))
			Expect(check.UpdatedAt).To(BeTemporally(">", user.UpdatedAt))
			Expect(check.Name).To(Equal(updateduser.Name))
			Expect(check.Roles).To(Equal(updateduser.Roles))

			By("expecting the password to not be changed")
			Expect(check.PasswordHash).To(Equal(user.PasswordHash))
		})

		It("can update the user's password but nothing else", func(ctx SpecContext) {
			user := &models.User{
				Name:         "first",
				Roles:        []string{"first"},
				PasswordHash: []byte("first"),
			}

			By("creating a new user")
			Expect(service.Create(ctx, user)).To(Succeed())
			Expect(user.ID).ToNot(BeEmpty())


			// Wait a bit so the UpdateAt timestamp can be increased
			time.Sleep(100 * time.Millisecond)

			By("updating the existing user")
			Expect(service.UpdatePassword(ctx, user.ID, []byte("second"))).To(Succeed())

			By("checking the new records")
			check, err := service.Get(ctx, user.ID)
			Expect(err).ToNot(HaveOccurred())

			Expect(check.ID).To(Equal(user.ID))
			Expect(check.UpdatedAt).To(BeTemporally(">", user.UpdatedAt))
			Expect(check.Name).To(Equal(user.Name))
			Expect(check.Roles).To(Equal(user.Roles))

			By("expecting the password to not be changed")
			Expect(check.PasswordHash).ToNot(Equal(user.PasswordHash))
		})
	})
})
