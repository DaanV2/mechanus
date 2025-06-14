package authenication_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/pkg/authenication"
	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/DaanV2/mechanus/server/pkg/storage"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
)

var _ = Describe("JwtService", func() {
	var (
		db         *database.DB
		jtiService *authenication.JTIService
		keyManager *authenication.KeyManager
		service    *authenication.JWTService
	)

	BeforeEach(func() {
		var err error

		db = util_test.CreateDatabase()
		jtiService = authenication.NewJTIService(db)
		dbstore := storage.DBStorage[*authenication.KeyData](db)
		keyManager, err = authenication.NewKeyManager(dbstore)
		Expect(err).ShouldNot(HaveOccurred())

		service = authenication.NewJWTService(jtiService, keyManager)
	})

	Describe("Create", func() {
		It("should create a JWT with a new JTI and reuse it while active", func(ctx SpecContext) {
			user := util_test.CreateUser()
			scope := "password"

			// Create first JWT
			token1, err := service.Create(ctx, user, scope)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token1).ToNot(BeEmpty())

			// Create second JWT, should reuse the same JTI
			token2, err := service.Create(ctx, user, scope)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(token2).ToNot(BeEmpty())

			// Parse both tokens and check JTI
			t1, err := service.Validate(ctx, token1)
			Expect(err).ShouldNot(HaveOccurred())
			claims1, ok := authenication.GetClaims(t1.Claims)
			Expect(ok).To(BeTrue())

			t2, err := service.Validate(ctx, token2)
			Expect(err).ShouldNot(HaveOccurred())
			claims2, ok := authenication.GetClaims(t2.Claims)
			Expect(ok).To(BeTrue())

			Expect(claims1.ID).To(Equal(claims2.ID))
			Expect(claims1.User.ID).To(Equal(user.ID))
		})

		It("should create a new JTI if the previous one is revoked", func(ctx SpecContext) {
			user := util_test.CreateUser()
			scope := "password"

			token1, err := service.Create(ctx, user, scope)
			Expect(err).ShouldNot(HaveOccurred())
			t1, err := service.Validate(ctx, token1)
			Expect(err).ShouldNot(HaveOccurred())
			claims1, ok := authenication.GetClaims(t1.Claims)
			Expect(ok).To(BeTrue())

			// Revoke the JTI
			_, err = jtiService.Revoke(ctx, claims1.ID)
			Expect(err).ShouldNot(HaveOccurred())

			// Create a new JWT, should have a new JTI
			token2, err := service.Create(ctx, user, scope)
			Expect(err).ShouldNot(HaveOccurred())
			t2, err := service.Validate(ctx, token2)
			Expect(err).ShouldNot(HaveOccurred())
			claims2, ok := authenication.GetClaims(t2.Claims)
			Expect(ok).To(BeTrue())

			Expect(claims2.ID).ToNot(Equal(claims1.ID))
		})
	})

	Describe("Validate", func() {
		It("should validate a valid JWT and return claims", func(ctx SpecContext) {
			user := util_test.CreateUser()
			scope := "password"
			token, err := service.Create(ctx, user, scope)
			Expect(err).ShouldNot(HaveOccurred())

			t, err := service.Validate(ctx, token)
			Expect(err).ShouldNot(HaveOccurred())
			claims, ok := authenication.GetClaims(t.Claims)
			Expect(ok).To(BeTrue())
			Expect(claims.User.ID).To(Equal(user.ID))
			Expect(claims.Scope).To(Equal(scope))
		})

		It("should fail validation if the JTI is revoked", func(ctx SpecContext) {
			user := util_test.CreateUser()
			scope := "password"
			token, err := service.Create(ctx, user, scope)
			Expect(err).ShouldNot(HaveOccurred())

			t, err := service.Validate(ctx, token)
			Expect(err).ShouldNot(HaveOccurred())
			claims, ok := authenication.GetClaims(t.Claims)
			Expect(ok).To(BeTrue())

			_, err = jtiService.Revoke(ctx, claims.ID)
			Expect(err).ShouldNot(HaveOccurred())

			_, err = service.Validate(ctx, token)
			Expect(err).To(HaveOccurred())
		})
	})
})
