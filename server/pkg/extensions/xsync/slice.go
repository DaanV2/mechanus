package xsync

import (
	"errors"
	"slices"
	"sync"
)

// Slice is a generic thread-safe slice with mutex-based locking.
type Slice[T any] struct {
	items []T
	lock  sync.Mutex
}

// NewSlice creates a new thread-safe Slice with the given items.
func NewSlice[T any](items ...T) *Slice[T] {
	return &Slice[T]{
		items: items,
		lock:  sync.Mutex{},
	}
}

// Append adds items to the slice in a thread-safe manner.
func (s *Slice[T]) Append(items ...T) *Slice[T] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.items = append(s.items, items...)

	return s
}

// Remove deletes all items matching the predicate. Returns true if any items were removed.
func (s *Slice[T]) Remove(predicate func(other T) bool) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	old := len(s.items)
	s.items = slices.DeleteFunc(s.items, predicate)

	return len(s.items) != old
}

// BorrowItems locks the slice and returns the items along with an unlock function.
// The caller must call the unlock function when done with the items.
func (s *Slice[T]) BorrowItems() (items []T, unlock func()) {
	s.lock.Lock()
	i := s.items
	un := func() {
		s.lock.Unlock()
	}

	return i, un
}

// RangeE iterates over all items and calls the function for each item.
// Returns the first error encountered, if any.
func (s *Slice[T]) RangeE(callfn func(item T) error) error {
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

// RangeErrorCollect loops over all the items and collect all the errors with [errors.Join]
func (s *Slice[T]) RangeErrorCollect(callfn func(item T) error) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	var err error

	for _, d := range s.items {
		err = errors.Join(callfn(d))

	}

	return err
}

// Range iterates over all items and calls the function for each item.
func (s *Slice[T]) Range(callfn func(item T)) {
	_ = s.RangeE(func(item T) error {
		callfn(item)

		return nil
	})
}
