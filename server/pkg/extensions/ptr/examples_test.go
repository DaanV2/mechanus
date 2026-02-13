package ptr_test

import (
	"fmt"
)

func ExampleTo() {
	// Convert an integer to a pointer
	num := 42
	numPtr := new(num)
	fmt.Printf("Value: %d, Pointer: %T\n", *numPtr, numPtr)
	// Output: Value: 42, Pointer: *int
}

func ExampleTo_string() {
	// Convert a string to a pointer
	str := "hello"
	strPtr := new(str)
	fmt.Printf("Value: %s, Pointer: %T\n", *strPtr, strPtr)
	// Output: Value: hello, Pointer: *string
}

func ExampleTo_struct() {
	type Person struct {
		Name string
		Age  int
	}

	person := Person{Name: "Alice", Age: 30}
	personPtr := new(person)
	fmt.Printf("Name: %s, Age: %d\n", personPtr.Name, personPtr.Age)
	// Output: Name: Alice, Age: 30
}
