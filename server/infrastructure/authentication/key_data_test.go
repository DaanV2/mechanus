package authentication_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xcrypto"
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
