package storage

import (
	"context"
	"iter"
)

// Identifiable is an interface for types that have a unique identifier.
type Identifiable interface {
	GetID() string
}

// Storage is a generic interface for storing and retrieving identifiable items.
type Storage[T Identifiable] interface {
	Get(ctx context.Context, id string) (T, error)
	Set(ctx context.Context, item T) error
	Keys(ctx context.Context) iter.Seq[string]
	Delete(ctx context.Context, item T) (bool, error)
}

// StorageProvider provides access to different storage locations.
type StorageProvider[T Identifiable] interface {
	AppStorage() (Storage[T], error)
	UserStorage() (Storage[T], error)
	StateStorage() (Storage[T], error)
}
