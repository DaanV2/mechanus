package modron

import (
	"slices"

	xsync "github.com/DaanV2/mechanus/server/pkg/extensions/sync"
)

type DuoDrone[T comparable] struct {
	tasks *xsync.Slice[T]
}

func NewDuoDrone[T comparable](items ...T) *DuoDrone[T] {
	return &DuoDrone[T]{
		tasks: xsync.NewSlice(items...),
	}
}

func (d *DuoDrone[T]) Add(tasks ...T) {
	d.tasks.Append(tasks...)
}

func (d *DuoDrone[T]) Remove(tasks ...T) bool {
	return d.tasks.Remove(func(other T) bool {
		return slices.Contains(tasks, other)
	})
}

func (d *DuoDrone[T]) RangeE(callfn func(item T) error) error {
	return d.tasks.WalkE(callfn)
}