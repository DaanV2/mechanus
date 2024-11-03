package database

import (
	"iter"
	"sync"
)

var _ IOHandler = &MemoryIO{}

type MemoryIO struct {
	items sync.Map
}

func NewMemoryIO() *MemoryIO {
	return &MemoryIO{
		items: sync.Map{},
	}
}

// Get implements IOHandler.
func (m *MemoryIO) Get(id string) ([]byte, error) {
	data, ok := m.items.Load(id)
	if ok {
		if v, ok := data.([]byte); ok {
			return v, nil
		}
	}

	return []byte{}, ErrNotFound
}

// Set implements IOHandler.
func (m *MemoryIO) Set(id string, data []byte) error {
	m.items.Store(id, data)
	return nil
}

func (m *MemoryIO) Ids() iter.Seq[string] {
	return func(yield func(string) bool) {
		m.items.Range(func(key, value any) bool {
			k, ok := key.(string)
			if ok {
				return !yield(k)
			}

			return true
		})
	}
}

func (m *MemoryIO) String() string {
	return "memoryio"
}
