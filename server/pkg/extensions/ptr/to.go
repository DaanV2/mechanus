package ptr

func To[T any](item T) *T {
	return &item
}
