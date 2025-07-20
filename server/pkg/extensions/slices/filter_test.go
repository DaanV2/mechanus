package xslices_test

import (
	xslices "github.com/DaanV2/mechanus/server/pkg/extensions/slices"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filter", func() {
	Context("with nil and empty inputs", func() {
		It("should handle nil slice", func() {
			var input []int = nil
			result := xslices.Filter(input, func(n int) bool {
				return true
			})
			Expect(result).To(BeNil())
		})

		It("should handle empty slice", func() {
			input := []int{}
			result := xslices.Filter(input, func(n int) bool {
				return true
			})
			Expect(result).To(BeEmpty())
			Expect(result).NotTo(BeNil())
		})
	})

	Context("with integers", func() {
		It("should exclude even numbers", func() {
			input := []int{1, 2, 3, 4, 5, 6}
			result := xslices.Filter(input, func(n int) bool {
				return n%2 == 0
			})
			Expect(result).To(Equal([]int{1, 3, 5}))
		})

		It("should exclude numbers greater than 3", func() {
			input := []int{1, 2, 3, 4, 5, 6}
			result := xslices.Filter(input, func(n int) bool {
				return n > 3
			})
			Expect(result).To(Equal([]int{1, 2, 3}))
		})

		It("should handle empty slice", func() {
			input := []int{}
			result := xslices.Filter(input, func(n int) bool {
				return n%2 == 0
			})
			Expect(result).To(BeEmpty())
		})

		It("should handle slice where no elements match predicate", func() {
			input := []int{1, 3, 5, 7}
			result := xslices.Filter(input, func(n int) bool {
				return n%2 == 0
			})
			Expect(result).To(Equal([]int{1, 3, 5, 7}))
		})
	})

	Context("with strings", func() {
		It("should exclude strings longer than 2", func() {
			input := []string{"a", "bb", "ccc", "dddd"}
			result := xslices.Filter(input, func(s string) bool {
				return len(s) > 2
			})
			Expect(result).To(Equal([]string{"a", "bb"}))
		})

		It("should exclude strings starting with 'a'", func() {
			input := []string{"apple", "banana", "avocado", "cherry"}
			result := xslices.Filter(input, func(s string) bool {
				return len(s) > 0 && s[0] == 'a'
			})
			Expect(result).To(Equal([]string{"banana", "cherry"}))
		})
	})

	Context("with custom types", func() {
		type Person struct {
			Name string
			Age  int
		}

		It("should exclude adults", func() {
			input := []Person{
				{Name: "Alice", Age: 25},
				{Name: "Bob", Age: 17},
				{Name: "Charlie", Age: 30},
				{Name: "David", Age: 16},
			}
			result := xslices.Filter(input, func(p Person) bool {
				return p.Age >= 18
			})
			Expect(result).To(Equal([]Person{
				{Name: "Bob", Age: 17},
				{Name: "David", Age: 16},
			}))
		})
	})
})
