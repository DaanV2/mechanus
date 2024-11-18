package storage

import (
	xerrors "github.com/DaanV2/mechanus/server/pkg/extensions/errors"
	"github.com/DaanV2/mechanus/server/pkg/generics"
)

type FirstQuery[T any] interface {
	First(predicate func(item T) bool) (T, error)
}

// First loops over all the data and looks for the item that matches the given predicate.
// returns a [xerrors.ErrNotExist] if nothing matched
func First[T any](storage Storage[T], predicate func(item T) bool) (T, error) {
	if v, ok := storage.(FirstQuery[T]); ok {
		return v.First(predicate)
	}

	for id := range storage.Ids() {
		item, err := storage.Get(id)
		if err != nil {
			return generics.Empty[T](), err
		}
		if predicate(item) {
			return item, nil
		}
	}

	return generics.Empty[T](), xerrors.ErrNotExist
}
