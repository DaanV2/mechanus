package xrand_test

import (
	xrand "github.com/DaanV2/mechanus/server/pkg/extensions/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Random ID Generation", func() {
	Context("MustID function", func() {
		It("should generate IDs of the requested length", func() {
			for l := range 64 {
				id := xrand.MustID(l)
				Expect(id).To(HaveLen(l))
			}
		})
	})
})
