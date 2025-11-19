package xslices

import "slices"

// Identifiable is an interface for types that have an ID.
type Identifiable[T any] interface {
	GetID() T
}

// ContainsID checks if any item in the slice has the given ID.
func ContainsID[S ~[]E, E Identifiable[U], U comparable](items S, id U) bool {
	return slices.ContainsFunc(items, func(item E) bool {
		return item.GetID() == id
	})
}

// CollectIDs extracts all IDs from the items in the slice.
func CollectIDs[S ~[]E, E Identifiable[U], U any](items S) []U {
	result := make([]U, 0, len(items))

	for _, item := range items {
		result = append(result, item.GetID())
	}

	return result
}

// AddIfMissing adds items to the slice if they don't already exist based on their ID.
func AddIfMissing[S ~[]E, E Identifiable[U], U comparable](items S, toadd ...E) S {
	for _, i := range toadd {
		id := i.GetID()
		if !ContainsID(items, id) {
			items = append(items, i)
		}
	}

	return items
}

// RemoveID removes items from the slice that match the IDs of items in toremove.
func RemoveID[S ~[]E, E Identifiable[U], U comparable](items S, toremove ...E) S {
	result := make(S, 0, len(items))

	for _, i := range items {
		if !ContainsID(toremove, i.GetID()) {
			result = append(result, i)
		}
	}

	return result
}
