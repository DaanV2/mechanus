package xslices_test

import (
	xslices "github.com/DaanV2/mechanus/server/pkg/extensions/slices"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Map", func() {
	Context("with nil and empty inputs", func() {
		It("should handle nil slice", func() {
			var input []int = nil
			result := xslices.Map(input, func(n int) int {
				return n * 2
			})
			Expect(result).To(BeEmpty())
			Expect(result).ToNot(BeNil())
		})

		It("should handle empty slice", func() {
			input := []int{}
			result := xslices.Map(input, func(n int) int {
				return n * 2
			})
			Expect(result).To(BeEmpty())
			Expect(result).ToNot(BeNil())
		})
	})

	Context("with integers", func() {
		It("should double each number", func() {
			input := []int{1, 2, 3}
			result := xslices.Map(input, func(n int) int {
				return n * 2
			})
			Expect(result).To(Equal([]int{2, 4, 6}))
		})

		It("should square each number", func() {
			input := []int{1, 2, 3, 4}
			result := xslices.Map(input, func(n int) int {
				return n * n
			})
			Expect(result).To(Equal([]int{1, 4, 9, 16}))
		})

		It("should handle empty slice", func() {
			input := []int{}
			result := xslices.Map(input, func(n int) int {
				return n * 2
			})
			Expect(result).To(BeEmpty())
		})
	})

	Context("with type transformations", func() {
		It("should convert integers to strings", func() {
			input := []int{1, 2, 3}
			result := xslices.Map(input, func(n int) string {
				return string(rune('A' - 1 + n))
			})
			Expect(result).To(Equal([]string{"A", "B", "C"}))
		})

		It("should convert strings to their lengths", func() {
			input := []string{"a", "bb", "ccc", "dddd"}
			result := xslices.Map(input, func(s string) int {
				return len(s)
			})
			Expect(result).To(Equal([]int{1, 2, 3, 4}))
		})
	})

	Context("with custom types", func() {
		type Person struct {
			Name string
			Age  int
		}

		type PersonDTO struct {
			FullName string
			IsAdult  bool
		}

		It("should transform Person to PersonDTO", func() {
			input := []Person{
				{Name: "Alice", Age: 25},
				{Name: "Bob", Age: 17},
				{Name: "Charlie", Age: 30},
			}
			result := xslices.Map(input, func(p Person) PersonDTO {
				return PersonDTO{
					FullName: p.Name,
					IsAdult:  p.Age >= 18,
				}
			})
			Expect(result).To(Equal([]PersonDTO{
				{FullName: "Alice", IsAdult: true},
				{FullName: "Bob", IsAdult: false},
				{FullName: "Charlie", IsAdult: true},
			}))
		})
	})
})
