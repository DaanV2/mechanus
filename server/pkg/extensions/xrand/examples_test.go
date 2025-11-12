package xrand_test

import (
	"fmt"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xrand"
)

func ExampleID() {
	// Generate a random ID of length 16
	id, err := xrand.ID(16)
	if err != nil {
		fmt.Println("Error:", err)

		return
	}

	fmt.Printf("ID length: %d\n", len(id))
	fmt.Printf("Is hex string: %t\n", isHexString(id))
	// Output:
	// ID length: 16
	// Is hex string: true
}

func ExampleID_oddLength() {
	// Works with odd lengths too
	id, err := xrand.ID(15)
	if err != nil {
		fmt.Println("Error:", err)

		return
	}

	fmt.Printf("ID length: %d\n", len(id))
	// Output:
	// ID length: 15
}

func ExampleMustID() {
	// Generate a random ID of length 32, panics on error
	id := xrand.MustID(32)

	fmt.Printf("ID length: %d\n", len(id))
	fmt.Printf("Is hex string: %t\n", isHexString(id))
	// Output:
	// ID length: 32
	// Is hex string: true
}

// Helper function for examples
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}

	return s != ""
}
