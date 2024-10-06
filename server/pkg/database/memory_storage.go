package database

import "sync"

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
