package must

// Do panics if err is not nil, otherwise returns value.
func Do[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}

	return value
}
