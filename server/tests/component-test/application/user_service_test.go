package application_test

import (
	"time"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/repositories"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("User Service", func() {

	var (
		db      *persistence.DB
		service *application.UserService
		repo    *repositories.UserRepository
	)

	BeforeEach(func(setupCtx SpecContext) {
		db = util_test.CreateDatabase(setupCtx)
		repo = repositories.NewUserRepository(db)
		service = application.NewUserService(repo)
	})

	Context("create", func() {
		It("can create a new user", func(ctx SpecContext) {
			user := &models.User{
				Username:     "gandalf",
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
			Expect(user.Username).To(Equal(u.Username))
			Expect(user.PasswordHash).To(Equal(u.PasswordHash))
			Expect(u.PasswordHash).ToNot(Equal([]byte("the-one-ring")))
			Expect(u.Roles).To(BeEquivalentTo([]string{"wizard"}))
		})
	})

	Context("Get", func() {
		It("can get a user by name", func(ctx SpecContext) {
			user := &models.User{
				Username: "gandalf",
			}

			By("creating a new user")
			Expect(service.Create(ctx, user)).To(Succeed())
			Expect(user.ID).ToNot(BeEmpty())

			By("trying to retrieve the user")
			u, err := service.FindByUsername(ctx, "gandalf")
			Expect(err).ToNot(HaveOccurred())

			Expect(user.ID).To(Equal(u.ID))
			Expect(user.Username).To(Equal(u.Username))
		})
	})

	Context("Update", func() {
		It("can update the user, but will skip the password and name", func(ctx SpecContext) {
			user := &models.User{
				Username:     "first",
				Roles:        []string{"first"},
				PasswordHash: []byte("first"),
			}

			By("creating a new user")
			Expect(service.Create(ctx, user)).To(Succeed())
			Expect(user.ID).ToNot(BeEmpty())

			updateduser := &models.User{
				Model:        user.Model,
				Username:     "second",
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
			Expect(check.Roles).To(Equal(updateduser.Roles))

			By("expecting the password and name to not be changed")
			Expect(check.PasswordHash).To(Equal(user.PasswordHash))
			Expect(check.Username).To(Equal(user.Username))
		})

		It("can update the user's password but nothing else", func(ctx SpecContext) {
			user := &models.User{
				Username:     "first",
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
			Expect(check.Username).To(Equal(user.Username))
			Expect(check.Roles).To(Equal(user.Roles))

			By("expecting the password to not be changed")
			Expect(check.PasswordHash).ToNot(Equal(user.PasswordHash))
		})
	})
})
