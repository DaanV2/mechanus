package must

import (
	"reflect"

	"github.com/charmbracelet/log"
)

// Call executes the provided function and logs a fatal error if it returns an error.
func Call[T any](call func() (T, error)) T {
	first, err := call()
	if err != nil {
		logger := log.Default()
		logger.Helper()
		logger.Fatal("couldn't perform call", "function", reflect.TypeOf(call).Name(), "error", err)
	}

	return first
}

// Must2 executes the provided function that returns two values and logs a fatal error if it returns an error.
func Must2[T any, U any](call func() (T, U, error)) (first T, second U) {
	first, second, err := call()
	if err != nil {
		logger := log.Default()
		logger.Helper()
		logger.Fatal("couldn't perform call", "function", reflect.TypeOf(call).Name(), "error", err)
	}

	return first, second
}
