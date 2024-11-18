package memory_storage

import (
	"iter"

	xerrors "github.com/DaanV2/mechanus/server/pkg/extensions/errors"
	xsync "github.com/DaanV2/mechanus/server/pkg/extensions/sync"
	"github.com/DaanV2/mechanus/server/pkg/generics"
)

type Storage[T any] struct {
	data *xsync.Map[string, T]
}

func NewStorage[T any]() *Storage[T] {
	return &Storage[T]{
		data: xsync.NewMap[string, T](),
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
	return s.data.Keys()
}

func (s *Storage[T]) get(id string) (T, error) {
	item, ok := s.data.Load(id)
	if ok {
		return item, nil
	}

	return generics.Empty[T](), xerrors.ErrNotExist
}

func (s *Storage[T]) set(id string, item T) error {
	s.data.Store(id, item)
	return nil
}
