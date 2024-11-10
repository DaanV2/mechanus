package memory_storage

import (
	"iter"
	"sync"

	xerrors "github.com/DaanV2/mechanus/server/pkg/extensions/errors"
)

type Storage[T any] struct {
	data sync.Map
}

func NewStorage[T any]() *Storage[T] {
	return &Storage[T]{
		data: sync.Map{},
	}
}

func (s *Storage[T]) Get(id string) (T, error) {
	return s.get(id)
}
func (s *Storage[T]) Set(id string, item T) error {
	return s.set(id, item)
}
func (s *Storage[T]) Has(id string) bool {
	_, ok := s.data.Load(id)
	return ok
}

func (s *Storage[T]) Ids() iter.Seq[string] {
	return func(yield func(string) bool) {
		s.data.Range(func(key, value any) bool {
			str, ok := key.(string)
			if !ok {
				return true //continue
			}

			return yield(str)
		})
	}
}

func (s *Storage[T]) get(id string) (T, error) {
	item, ok := s.data.Load(id)
	if ok {
		v, ok := item.(T)
		if ok {
			return v, nil
		}
	}

	var empty T
	return empty, xerrors.ErrNotExist
}

func (s *Storage[T]) set(id string, item T) error {
	s.data.Store(id, item)
	return nil
}
