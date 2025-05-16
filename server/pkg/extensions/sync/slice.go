package xsync

import (
	"slices"
	"sync"
)

type Slice[T any] struct {
	items []T
	lock  sync.Mutex
}

func NewSlice[T any](items ...T) *Slice[T] {
	return &Slice[T]{
		items: items,
		lock:  sync.Mutex{},
	}
}

func (s *Slice[T]) Append(items ...T) *Slice[T] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.items = append(s.items, items...)

	return s
}

func (s *Slice[T]) Remove(predicate func(other T) bool) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	old := len(s.items)
	s.items = slices.DeleteFunc(s.items, predicate)

	return len(s.items) != old
}

func (s *Slice[T]) Items() (items []T, unlock func()) {
	s.lock.Lock()
	i := s.items
	un := func() {
		s.lock.Unlock()
	}

	return i, un
}

func (s *Slice[T]) WalkE(callfn func(item T) error) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, d := range s.items {
		err := callfn(d)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Slice[T]) Walk(callfn func(item T)) {
	_ = s.WalkE(func(item T) error {
		callfn(item)

		return nil
	})
}
