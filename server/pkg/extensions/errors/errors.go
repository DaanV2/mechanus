package xerrors

import (
	"errors"
	"reflect"

	"github.com/charmbracelet/log"
)

var (
	ErrNotExist = errors.New("item does not exist")
)

func Must[T any](call func() (T, error)) T {
	first, err := call()
	if err != nil {
		logger := log.Default()
		logger.Helper()
		logger.Fatal("couldn't perform call", "function", reflect.TypeOf(call).Name(), "error", err)
	}

	return first
}

func Must2[T any, U any](call func() (T, U, error)) (first T, second U) {
	first, second, err := call()
	if err != nil {
		logger := log.Default()
		logger.Helper()
		logger.Fatal("couldn't perform call", "function", reflect.TypeOf(call).Name(), "error", err)
	}

	return first, second
}