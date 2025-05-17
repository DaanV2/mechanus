package xerrors

import (
	"errors"
	"reflect"

	"github.com/charmbracelet/log"
)

var (
	ErrNotExist = errors.New("item does not exist")
)

func Must[T any](call func() (v T, err error)) T {
	v, err := call()
	if err != nil {
		logger := log.Default()
		logger.Helper()
		logger.Fatal("couldn't perform call", "function", reflect.TypeOf(call).Name(), "error", err)
	}

	return v
}

func Must2[T any, U any](call func() (a T, b U, err error)) (T, U) {
	a, b, err := call()
	if err != nil {
		logger := log.Default()
		logger.Helper()
		logger.Fatal("couldn't perform call", "function", reflect.TypeOf(call).Name(), "error", err)
	}

	return a, b
}