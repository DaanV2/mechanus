package bools_test

import (
	"fmt"

	"github.com/DaanV2/mechanus/server/pkg/extensions/bools"
)

func ExampleAnd() {
	result1 := bools.And(true, true, true)
	result2 := bools.And(true, false, true)
	result3 := bools.And(true, true)

	fmt.Println(result1)
	fmt.Println(result2)
	fmt.Println(result3)
	// Output:
	// true
	// false
	// true
}

func ExampleAnd_empty() {
	// Empty input returns true
	result := bools.And()
	fmt.Println(result)
	// Output: true
}

func ExampleOr() {
	result1 := bools.Or(false, false, false)
	result2 := bools.Or(false, true, false)
	result3 := bools.Or(true, false)

	fmt.Println(result1)
	fmt.Println(result2)
	fmt.Println(result3)
	// Output:
	// false
	// true
	// true
}

func ExampleOr_empty() {
	// Empty input returns false
	result := bools.Or()
	fmt.Println(result)
	// Output: false
}
