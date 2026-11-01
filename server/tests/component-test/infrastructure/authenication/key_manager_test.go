package authentication_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xcrypto"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
)

var _ = Describe("KeyData", func() {
	It("round trips through PEM", func() {
		rsa, err := xcrypto.GenerateRSAKeys()
		Expect(err).ShouldNot(HaveOccurred())
		created := time.Now().Truncate(time.Second)
		key := authentication.NewKeyData(rsa.ID(), rsa.Private(), created)

		text, err := key.MarshalText()
		Expect(err).ShouldNot(HaveOccurred())

		var result authentication.KeyData
		Expect(result.UnmarshalText(text)).To(Succeed())
		Expect(result.GetID()).To(Equal(key.GetID()))
		Expect(result.Created()).To(BeTemporally("==", created))
		Expect(result.Public()).To(Equal(key.Public()))
	})

	It("reads keys stored without a creation time", func() {
		rsa, err := xcrypto.GenerateRSAKeys()
		Expect(err).ShouldNot(HaveOccurred())
		text, err := authentication.NewKeyData(rsa.ID(), rsa.Private(), time.Time{}).MarshalText()
		Expect(err).ShouldNot(HaveOccurred())

		var result authentication.KeyData
		Expect(result.UnmarshalText(text)).To(Succeed())
		Expect(result.Created().IsZero()).To(BeTrue())
	})

	It("rejects data that isn't PEM", func() {
		var result authentication.KeyData
		Expect(result.UnmarshalText([]byte("not a key"))).ToNot(Succeed())
	})
})

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
