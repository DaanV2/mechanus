package generics_test

import (
	"testing"

	"github.com/DaanV2/mechanus/server/pkg/generics"
	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	A int32
	B int32
	C int32
}

func Test_NameOf(t *testing.T) {
	assert.Equal(t, "bool", generics.NameOf[bool]())
	assert.Equal(t, "uint", generics.NameOf[uint]())
	assert.Equal(t, "uint8", generics.NameOf[uint8]())
	assert.Equal(t, "uint16", generics.NameOf[uint16]())
	assert.Equal(t, "uint32", generics.NameOf[uint32]())
	assert.Equal(t, "uint64", generics.NameOf[uint64]())
	assert.Equal(t, "int", generics.NameOf[int]())
	assert.Equal(t, "int8", generics.NameOf[int8]())
	assert.Equal(t, "int16", generics.NameOf[int16]())
	assert.Equal(t, "int32", generics.NameOf[int32]())
	assert.Equal(t, "int64", generics.NameOf[int64]())
	assert.Equal(t, "float32", generics.NameOf[float32]())
	assert.Equal(t, "float64", generics.NameOf[float64]())
	assert.Equal(t, "complex128", generics.NameOf[complex128]())
	assert.Equal(t, "uint8", generics.NameOf[byte]())
	assert.Equal(t, "int32", generics.NameOf[rune]())
	assert.Empty(t, generics.NameOf[any]())
	assert.Equal(t, "TestCase", generics.NameOf[TestCase]())
}

func Test_SizeOf(t *testing.T) {
	assert.Equal(t, generics.SizeOf[bool](), uintptr(1))
	assert.Equal(t, generics.SizeOf[uint8](), uintptr(1))
	assert.Equal(t, generics.SizeOf[uint16](), uintptr(2))
	assert.Equal(t, generics.SizeOf[uint32](), uintptr(4))
	assert.Equal(t, generics.SizeOf[uint64](), uintptr(8))
	assert.Equal(t, generics.SizeOf[int8](), uintptr(1))
	assert.Equal(t, generics.SizeOf[int16](), uintptr(2))
	assert.Equal(t, generics.SizeOf[int32](), uintptr(4))
	assert.Equal(t, generics.SizeOf[int64](), uintptr(8))
	assert.Equal(t, generics.SizeOf[float32](), uintptr(4))
	assert.Equal(t, generics.SizeOf[float64](), uintptr(8))
	assert.Equal(t, generics.SizeOf[complex128](), uintptr(16))
	assert.Equal(t, generics.SizeOf[byte](), uintptr(1))
	assert.Equal(t, generics.SizeOf[rune](), uintptr(4))
	assert.Equal(t, generics.SizeOf[any](), uintptr(16))
	assert.Equal(t, generics.SizeOf[TestCase](), uintptr(12))
	assert.Equal(t, generics.SizeOf[*TestCase](), uintptr(8))
}
