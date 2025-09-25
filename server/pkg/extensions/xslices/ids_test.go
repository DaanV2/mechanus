package xslices_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xslices"
)

type testItem struct {
	id  int
	val string
}

func (t testItem) GetID() int { return t.id }

var _ = Describe("xslices Identifiable helpers", func() {

	var (
		items []testItem
	)

	BeforeEach(func() {
		items = []testItem{
			{id: 1, val: "a"},
			{id: 2, val: "b"},
			{id: 3, val: "c"},
		}
	})

	Describe("ContainsID", func() {
		It("returns true if ID exists", func() {
			Expect(xslices.ContainsID(items, 2)).To(BeTrue())
		})
		It("returns false if ID does not exist", func() {
			Expect(xslices.ContainsID(items, 99)).To(BeFalse())
		})
	})

	Describe("CollectIDs", func() {
		It("collects all IDs", func() {
			Expect(xslices.CollectIDs(items)).To(ConsistOf(1, 2, 3))
		})
	})

	Describe("AddIfMissing", func() {
		It("adds only missing items", func() {
			newItem := testItem{id: 4, val: "d"}
			result := xslices.AddIfMissing(items, newItem)
			Expect(result).To(ContainElement(newItem))
			Expect(result).To(HaveLen(4))
		})
		It("does not add duplicates", func() {
			dup := testItem{id: 2, val: "b"}
			result := xslices.AddIfMissing(items, dup)
			Expect(result).To(HaveLen(3))
		})
	})

	Describe("Remove", func() {
		It("removes items with matching IDs", func() {
			toRemove := testItem{id: 2}
			result := xslices.RemoveID(items, toRemove)
			Expect(result).To(HaveLen(2))
			Expect(xslices.ContainsID(result, 2)).To(BeFalse())
		})
		It("does not remove items with non-matching IDs", func() {
			toRemove := testItem{id: 99}
			result := xslices.RemoveID(items, toRemove)
			Expect(result).To(HaveLen(3))
		})
	})
})
