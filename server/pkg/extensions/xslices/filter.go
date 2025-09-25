package xslices

// Filter removes elements from a slice based on a predicate function.
// It returns a new slice containing only the elements that do not satisfy the predicate.
func Filter[S ~[]E, E any](items S, predicate func(E) bool) S {
	result := make(S, 0, len(items))

	for _, item := range items {
		if !predicate(item) {
			result = append(result, item)
		}
	}

	return result
}
