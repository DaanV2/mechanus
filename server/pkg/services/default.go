package services

import "iter"

type ObjectService[T any] interface {
	Get(id string) (T, error)
	Update(object T) error
	Create(object T) error
	Iterator() iter.Seq2[string, T]
}
