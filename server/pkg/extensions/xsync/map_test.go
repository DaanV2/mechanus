package xsync_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xsync"
)

var _ = Describe("Map", func() {
	Describe("Swap", func() {
		It("returns the zero value for a new key", func() {
			m := xsync.NewMap[string, *int]()
			value := 1

			previous, loaded := m.Swap("a", &value)
			Expect(loaded).To(BeFalse())
			Expect(previous).To(BeNil())

			stored, ok := m.Load("a")
			Expect(ok).To(BeTrue())
			Expect(stored).To(Equal(&value))
		})

		It("returns the previous value for an existing key", func() {
			m := xsync.NewMap[string, int]()
			m.Store("a", 1)

			previous, loaded := m.Swap("a", 2)
			Expect(loaded).To(BeTrue())
			Expect(previous).To(Equal(1))
		})
	})
})
