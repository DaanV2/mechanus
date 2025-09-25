package xslices

import "slices"

type Identifiable[T any] interface {
	GetID() T
}

func ContainsID[S ~[]E, E Identifiable[U], U comparable](items S, id U) bool {
	return slices.ContainsFunc(items, func(item E) bool {
		return item.GetID() == id
	})
}

func CollectIDs[S ~[]E, E Identifiable[U], U any](items S) []U {
	result := make([]U, 0, len(items))

	for _, item := range items {
		result = append(result, item.GetID())
	}

	return result
}

func AddIfMissing[S ~[]E, E Identifiable[U], U comparable](items S, toadd ...E) S {
	for _, i := range toadd {
		id := i.GetID()
		if !ContainsID(items, id) {
			items = append(items, i)
		}
	}

	return items
}

func RemoveID[S ~[]E, E Identifiable[U], U comparable](items S, toremove ...E) S {
	result := make(S, 0, len(items))

	for _, i := range items {
		if !ContainsID(toremove, i.GetID()) {
			result = append(result, i)
		}
	}

	return result
}
