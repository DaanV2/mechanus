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
	assert.Equal(t, generics.NameOf[bool](), "bool")
	assert.Equal(t, generics.NameOf[uint](), "uint")
	assert.Equal(t, generics.NameOf[uint8](), "uint8")
	assert.Equal(t, generics.NameOf[uint16](), "uint16")
	assert.Equal(t, generics.NameOf[uint32](), "uint32")
	assert.Equal(t, generics.NameOf[uint64](), "uint64")
	assert.Equal(t, generics.NameOf[int](), "int")
	assert.Equal(t, generics.NameOf[int8](), "int8")
	assert.Equal(t, generics.NameOf[int16](), "int16")
	assert.Equal(t, generics.NameOf[int32](), "int32")
	assert.Equal(t, generics.NameOf[int64](), "int64")
	assert.Equal(t, generics.NameOf[float32](), "float32")
	assert.Equal(t, generics.NameOf[float64](), "float64")
	assert.Equal(t, generics.NameOf[complex128](), "complex128")
	assert.Equal(t, generics.NameOf[byte](), "uint8")
	assert.Equal(t, generics.NameOf[rune](), "int32")
	assert.Equal(t, generics.NameOf[any](), "")
	assert.Equal(t, generics.NameOf[TestCase](), "TestCase")
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
