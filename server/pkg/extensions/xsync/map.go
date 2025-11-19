package xsync

import (
	"iter"
	"sync"

	"github.com/daanv2/go-kit/generics"
)

// Map is a generic thread-safe map wrapper around sync.Map.
type Map[K, V any] struct {
	data sync.Map
}

// NewMap creates a new thread-safe Map.
func NewMap[K, V any]() *Map[K, V] {
	return &Map[K, V]{
		data: sync.Map{},
	}
}

// Delete removes the value associated with the given key from the map.
func (m *Map[K, V]) Delete(key K) {
	m.data.Delete(key)
}

// CompareAndDelete deletes the entry for key if its value is equal to old.
func (m *Map[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return m.data.CompareAndDelete(key, old)
}

// Load returns the value stored in the map for a key, or a zero value if no value is present.
// The ok result indicates whether value was found in the map.
func (m *Map[K, V]) Load(key K) (V, bool) {
	v, ok := m.data.Load(key)
	if !ok {
		return generics.Empty[V](), ok
	}

	value, ok := v.(V)

	return value, ok
}

// Store sets the value for a key.
func (m *Map[K, V]) Store(key K, value V) {
	m.data.Store(key, value)
}

// Swap swaps the value for a key and returns the previous value if any.
// The loaded result reports whether the key was present.
func (m *Map[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	v, loaded := m.data.Swap(key, value)

	return v.(V), loaded
}

// StoreAll stores all key-value pairs from the provided iterator into the map.
func (m *Map[K, V]) StoreAll(items iter.Seq2[K, V]) {
	for k, v := range items {
		m.Store(k, v)
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value. The loaded result is true if the value was loaded, false if stored.
func (ht *Map[K, V]) LoadOrStore(key K, value V) (result V, loaded bool) {
	var tmp any
	tmp, loaded = ht.data.LoadOrStore(key, value)
	if tmp == nil {
		return generics.Empty[V](), loaded
	}

	result, ok := tmp.(V)
	if !ok {
		return generics.Empty[V](), false
	}

	return result, loaded
}

// Clear removes all entries from the map.
func (m *Map[K, V]) Clear() {
	m.data.Clear()
}

// Range calls f sequentially for each key and value present in the map.
func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.data.Range(func(key, value any) bool {
		k, kok := key.(K)
		v, vok := value.(V)

		if kok && vok {
			return f(k, v)
		}

		return true
	})
}

// Keys returns an iterator over the keys in the map.
func (m *Map[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		m.data.Range(func(key, value any) bool {
			k, ok := key.(K)
			if ok && !yield(k) {
				return false
			}

			return true
		})
	}
}

// Items returns an iterator over the key-value pairs in the map.
func (m *Map[K, V]) Items() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		m.data.Range(func(key, value any) bool {
			k, kok := key.(K)
			v, vok := value.(V)

			if kok && vok {
				if !yield(k, v) {
					return false
				}
			}

			return true
		})
	}
}
