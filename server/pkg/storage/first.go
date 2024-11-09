package storage

type FirstQuery[T any] interface {
	First(predicate func(item T) bool) (T, error)
}

// First loops over all the data and looks for the item that matches the given predicate
func First[T any](storage Storage[T], predicate func(item T) bool) (T, error) {
	if v, ok := storage.(FirstQuery[T]); ok {
		return v.First(predicate)
	}

	var empty T
	for id := range storage.Ids() {
		item, err := storage.Get(id)
		if err != nil {
			return empty, err
		}
		if predicate(item) {
			return item, nil
		}
	}

	return empty, nil
}
