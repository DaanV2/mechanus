package file_storage

import (
	"encoding"
	"encoding/json"
	"iter"
	"path/filepath"

	"github.com/DaanV2/mechanus/server/pkg/generics"
	"github.com/DaanV2/mechanus/server/pkg/storage"
	"github.com/charmbracelet/log"
)

type Storage[T any] struct {
	name   string
	raw    *RawStorage
	logger *log.Logger
}

func NewStorage[T any](folder string) *Storage[T] {
	t := generics.NameOf[T]()
	f := filepath.Join(folder)

	return &Storage[T]{
		name:   t,
		raw:    NewRawStorage(f),
		logger: log.With("storage", t),
	}
}

func (s *Storage[T]) Ids() iter.Seq[string] {
	return s.raw.Ids()
}

func (s *Storage[T]) Get(id string) (T, error) {
	s.logger.Debugf("retrieving '%s' from storage '%s'", id, s.name)

	var result T
	data, err := s.raw.Get(id)
	if err != nil {
		return result, err
	}

	if v, ok := interface{}(result).(encoding.BinaryUnmarshaler); ok {
		err = v.UnmarshalBinary(data)
	} else if v, ok := interface{}(result).(encoding.TextUnmarshaler); ok {
		err = v.UnmarshalText(data)
	} else {
		err = json.Unmarshal(data, &result)
	}

	return result, err
}

func (s *Storage[T]) Set(id string, item T) error {
	s.logger.Debugf("setting '%s' to storage '%s'", id, s.name)
	var (
		data []byte
		err  error
	)

	if v, ok := interface{}(item).(encoding.BinaryMarshaler); ok {
		data, err = v.MarshalBinary()
	} else if v, ok := interface{}(item).(encoding.TextMarshaler); ok {
		data, err = v.MarshalText()
	} else {
		data, err = json.Marshal(item)
	}
	if err != nil {
		return err
	}

	return s.raw.Set(id, data)
}

// First returns the item that matches the given predicate first, returns [ErrNotFound] is nothing is found
func (Storage *Storage[T]) First(predicate func(item T) bool) (T, error) {
	var empty T
	for id := range Storage.Ids() {
		v, err := Storage.Get(id)
		if err != nil {
			return empty, err
		}
		if predicate(v) {
			return v, nil
		}
	}

	return empty, storage.ErrNotExist
}
