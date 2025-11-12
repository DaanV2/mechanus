package xstrings_test

import (
	"fmt"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xstrings"
)

func ExampleFirstNotEmpty() {
	result := xstrings.FirstNotEmpty("", "", "hello", "world")
	fmt.Println(result)
	// Output: hello
}

func ExampleFirstNotEmpty_allEmpty() {
	result := xstrings.FirstNotEmpty("", "", "")
	fmt.Println(result)
	// Output:
}

func ExampleFirstNotEmpty_firstNotEmpty() {
	result := xstrings.FirstNotEmpty("first", "second", "third")
	fmt.Println(result)
	// Output: first
}
