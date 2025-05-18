package xslices

type Identifiable[T any] interface {
	GetID() T
}

func CollectIDs[S ~[]E, E Identifiable[U], U any](items S) []U {
	result := make([]U, 0, len(items))

	for _, item := range items {
		result = append(result, item.GetID())
	}

	return result
}