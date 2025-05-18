package generics_test

import (
	"github.com/DaanV2/mechanus/server/pkg/generics"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type TestCase struct {
	A int32
	B int32
	C int32
}

var _ = Describe("Generics", func() {
	Describe("NameOf", func() {
		DescribeTable("should return the correct type names",
			func(expected string, fn func() string) {
				Expect(fn()).To(Equal(expected))
			},
			Entry("bool", "bool", generics.NameOf[bool]),
			Entry("uint", "uint", generics.NameOf[uint]),
			Entry("uint8", "uint8", generics.NameOf[uint8]),
			Entry("uint16", "uint16", generics.NameOf[uint16]),
			Entry("uint32", "uint32", generics.NameOf[uint32]),
			Entry("uint64", "uint64", generics.NameOf[uint64]),
			Entry("int", "int", generics.NameOf[int]),
			Entry("int8", "int8", generics.NameOf[int8]),
			Entry("int16", "int16", generics.NameOf[int16]),
			Entry("int32", "int32", generics.NameOf[int32]),
			Entry("int64", "int64", generics.NameOf[int64]),
			Entry("float32", "float32", generics.NameOf[float32]),
			Entry("float64", "float64", generics.NameOf[float64]),
			Entry("complex128", "complex128", generics.NameOf[complex128]),
			Entry("byte (alias for uint8)", "uint8", generics.NameOf[byte]),
			Entry("rune (alias for int32)", "int32", generics.NameOf[rune]),
			Entry("any", "", generics.NameOf[any]),
			Entry("TestCase", "TestCase", generics.NameOf[TestCase]),
		)
	})

	Describe("SizeOf", func() {
		DescribeTable("should return the correct size",
			func(expected int, fn func() uintptr) {
				Expect(fn()).To(BeEquivalentTo(expected))
			},
			Entry("bool", 1, generics.SizeOf[bool]),
			Entry("uint8", 1, generics.SizeOf[uint8]),
			Entry("uint16", 2, generics.SizeOf[uint16]),
			Entry("uint32", 4, generics.SizeOf[uint32]),
			Entry("uint64", 8, generics.SizeOf[uint64]),
			Entry("int8", 1, generics.SizeOf[int8]),
			Entry("int16", 2, generics.SizeOf[int16]),
			Entry("int32", 4, generics.SizeOf[int32]),
			Entry("int64", 8, generics.SizeOf[int64]),
			Entry("float32", 4, generics.SizeOf[float32]),
			Entry("float64", 8, generics.SizeOf[float64]),
			Entry("complex128", 16, generics.SizeOf[complex128]),
			Entry("byte (alias for uint8)", 1, generics.SizeOf[byte]),
			Entry("rune (alias for int32)", 4, generics.SizeOf[rune]),
			Entry("any", 16, generics.SizeOf[any]),
			Entry("TestCase", 12, generics.SizeOf[TestCase]),
			Entry("*TestCase", 8, generics.SizeOf[*TestCase]),
		)
	})
})
