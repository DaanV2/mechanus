package storage

import "iter"

type Storage[T any] interface {
	Get(id string) (T, error)
	Set(id string, item T) error
	Has(id string) bool
	Ids() iter.Seq[string]
}
