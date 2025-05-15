package storage

import "context"

type Identifiable interface {
	GetID() string
}

type Storage[T Identifiable] interface {
	Get(ctx context.Context, id string) (T, error)
	Set(ctx context.Context, item T) error
	Delete(ctx context.Context, item T) (bool, error)
}

type StorageProvider[T Identifiable] interface {
	AppStorage() (Storage[T], error)
	UserStorage() (Storage[T], error)
	StateStorage() (Storage[T], error)
}
