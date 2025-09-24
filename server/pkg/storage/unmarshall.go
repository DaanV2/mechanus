package storage

import (
	"errors"
	"reflect"

	xencoding "github.com/DaanV2/mechanus/server/pkg/extensions/encoding"
	"github.com/daanv2/go-kit/generics"
)

// unmarshallGeneric helps with the fact that its unknown if T is a struct or a pointer to a struct, making unmarshalling difficult
func unmarshallGeneric[T any](data []byte) (T, error) {
	var result T
	var unmarshalTarget any
	typeOfT := reflect.TypeFor[T]()
	if typeOfT.Kind() == reflect.Ptr {
		// T is a pointer type, allocate a new value of the element type
		result = reflect.New(typeOfT.Elem()).Interface().(T)
		unmarshalTarget = result
	} else {
		// T is a value type, pass its address
		unmarshalTarget = &result
	}

	err := xencoding.Unmarshal(data, unmarshalTarget)
	if err != nil {
		return generics.Empty[T](), err
	}

	if v, ok := unmarshalTarget.(T); ok {
		return v, nil
	}
	if v, ok := unmarshalTarget.(*T); ok {
		return *v, nil
	}

	return generics.Empty[T](), errors.New("i have no idea what this is")
}
