package authentication_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
	util_test "github.com/DaanV2/mechanus/server/tests/util"
)

var _ = Describe("KeyManager", func() {
	var provider storage.StorageProvider[*authentication.KeyData]

	BeforeEach(func(setupCtx SpecContext) {
		provider = storage.DBStorage[*authentication.KeyData](util_test.CreateDatabase(setupCtx))
	})

	newManager := func() *authentication.KeyManager {
		GinkgoHelper()
		manager, err := authentication.NewKeyManager(provider)
		Expect(err).ShouldNot(HaveOccurred())

		return manager
	}

	It("reuses the signing key", func(ctx SpecContext) {
		manager := newManager()
		first, err := manager.GetSigningKey(ctx)
		Expect(err).ShouldNot(HaveOccurred())
		second, err := manager.GetSigningKey(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(second.GetID()).To(Equal(first.GetID()))
	})

	It("signs with the newest key once there are several", func(ctx SpecContext) {
		manager := newManager()
		_, err := manager.New(ctx)
		Expect(err).ShouldNot(HaveOccurred())
		time.Sleep(time.Millisecond * 5)
		newest, err := manager.New(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		for range 5 {
			key, gerr := manager.GetSigningKey(ctx)
			Expect(gerr).ShouldNot(HaveOccurred())
			Expect(key.GetID()).To(Equal(newest.GetID()))
		}
	})

	It("loads stored keys on AfterInitialize", func(ctx SpecContext) {
		key, err := newManager().GetSigningKey(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		manager := newManager()
		Expect(manager.AfterInitialize(ctx)).To(Succeed())

		loaded, err := manager.GetSigningKey(ctx)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(loaded.GetID()).To(Equal(key.GetID()))
	})

	It("errors on an unknown kid", func(ctx SpecContext) {
		_, err := newManager().Get(ctx, "does-not-exist")
		Expect(err).To(HaveOccurred())
	})

	It("rotates to a new key and retires keys whose tokens all expired", func(ctx SpecContext) {
		manager := newManager().WithRotation(time.Millisecond*300, time.Second)
		first, err := manager.GetSigningKey(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		time.Sleep(time.Millisecond * 350)
		second, err := manager.GetSigningKey(ctx)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(second.GetID()).ToNot(Equal(first.GetID()))

		// Still within the token lifetime, tokens signed by the first key are still valid
		_, err = manager.Get(ctx, first.GetID())
		Expect(err).ShouldNot(HaveOccurred())

		time.Sleep(time.Second)
		third, err := manager.GetSigningKey(ctx)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(third.GetID()).ToNot(Equal(second.GetID()))

		_, err = manager.Get(ctx, first.GetID())
		Expect(err).To(HaveOccurred(), "the first key should have been retired")
	})
})
