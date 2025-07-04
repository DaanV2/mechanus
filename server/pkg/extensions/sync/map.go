package xsync

import (
	"iter"
	"sync"

	"github.com/daanv2/go-kit/generics"
)

type Map[K, V any] struct {
	data sync.Map
}

func NewMap[K, V any]() *Map[K, V] {
	return &Map[K, V]{
		data: sync.Map{},
	}
}

func (m *Map[K, V]) Load(key K) (V, bool) {
	v, ok := m.data.Load(key)
	if !ok {
		return generics.Empty[V](), ok
	}

	value, ok := v.(V)

	return value, ok
}

func (m *Map[K, V]) Store(key K, value V) {
	m.data.Store(key, value)
}

func (m *Map[K, V]) StoreAll(items iter.Seq2[K, V]) {
	for k, v := range items {
		m.Store(k, v)
	}
}

func (m *Map[K, V]) Clear() {
	m.data.Clear()
}

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
