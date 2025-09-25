package xcrypto_test

import (
	"github.com/DaanV2/mechanus/server/pkg/extensions/xcrypto"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RSA", func() {
	Context("Generation", func() {
		It("should be able to generate a pair of keys for RSA", func() {
			key, err := xcrypto.GenerateRSAKeys()
			Expect(err).ToNot(HaveOccurred())

			Expect(key.ID()).ToNot(BeEmpty())
		})
	})
})
