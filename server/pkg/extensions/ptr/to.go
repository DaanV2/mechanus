package ptr

// To returns a pointer to the given item. Useful for converting values to pointers inline.
//
//go:fix inline
func To[T any](item T) *T {
	return new(item)
}
