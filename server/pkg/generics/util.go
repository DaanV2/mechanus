package generics

import "reflect"

func NameOf[T any]() string {
	return reflect.TypeFor[T]().Name()
}

func SizeOf[T any]() uintptr {
	return reflect.TypeFor[T]().Size()
}

func Empty[T any]() T {
	var empty T

	return empty
}
