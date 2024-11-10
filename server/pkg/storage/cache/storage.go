package cache_storage

import (
	"errors"
	"iter"

	"github.com/DaanV2/mechanus/server/pkg/generics"
	"github.com/DaanV2/mechanus/server/pkg/storage"
	memory_storage "github.com/DaanV2/mechanus/server/pkg/storage/memory"
	"github.com/charmbracelet/log"
)

var _ storage.Storage[any] = &Cache[any]{}
var _ storage.FirstQuery[any] = &Cache[any]{}

type Cache[T any] struct {
	memory *memory_storage.Storage[T]
	base   storage.Storage[T]
}

func NewCache[T any](base storage.Storage[T]) *Cache[T] {
	// TODO add auto cleaning
	return &Cache[T]{
		memory: memory_storage.NewStorage[T](),
		base:   base,
	}
}

// Get implements storage.Storage.
func (c *Cache[T]) Get(id string) (T, error) {
	var empty T
	item, err := c.memory.Get(id)
	if err == nil {
		return item, nil
	}

	item, err = c.base.Get(id)
	if err != nil {
		return empty, nil
	}
	err = c.memory.Set(id, item) // Attempt to set
	if err != nil {
		log.Warn("couldn't store item in memory", "id", id, "type", generics.NameOf[T]())
	}

	return item, nil
}

// Has implements storage.Storage.
func (c *Cache[T]) Has(id string) bool {
	return c.memory.Has(id) || c.base.Has(id)
}

// Ids implements storage.Storage.
func (c *Cache[T]) Ids() iter.Seq[string] {
	return c.base.Ids()
}

// Set implements storage.Storage.
func (c *Cache[T]) Set(id string, item T) error {
	return errors.Join(
		c.memory.Set(id, item),
		c.base.Set(id, item),
	)
}

// First implements storage.FirstQuery.
func (c *Cache[T]) First(predicate func(item T) bool) (T, error) {
	panic("unimplemented")
}
