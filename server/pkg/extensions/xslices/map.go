package xslices

// Map transforms each element in the slice using the provided transform function.
func Map[S ~[]E, E comparable, U any](items S, transform func(E) U) []U {
	result := make([]U, 0, len(items))

	for _, item := range items {
		result = append(result, transform(item))
	}

	return result
}
