package storage

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xformat"
	"github.com/daanv2/go-kit/generics"

	"gorm.io/gorm/clause"
)

// DBStorage creates a database-based storage provider for the given type, keyed by the type name.
func DBStorage[T Identifiable](db *persistence.DB) StorageProvider[T] {
	return NamedDBStorage[T](db, generics.NameOf[T]())
}

// NamedDBStorage creates a database-based storage provider for the given type, keyed by the given name.
func NamedDBStorage[T Identifiable](db *persistence.DB, name string) StorageProvider[T] {
	return &dbProvider[T]{db: db, name: name}
}

type dbProvider[T Identifiable] struct {
	db   *persistence.DB
	name string
}

type dbStorage[T Identifiable] struct {
	db     *persistence.DB
	prefix string
}

// AppStorage implements StorageProvider.
func (d *dbProvider[T]) AppStorage() (Storage[T], error) {
	return &dbStorage[T]{
		db:     d.db,
		prefix: fmt.Sprintf("app/%s/", d.name),
	}, nil
}

// StateStorage implements StorageProvider.
func (d *dbProvider[T]) StateStorage() (Storage[T], error) {
	return &dbStorage[T]{
		db:     d.db,
		prefix: fmt.Sprintf("state/%s/", d.name),
	}, nil
}

// UserStorage implements StorageProvider.
func (d *dbProvider[T]) UserStorage() (Storage[T], error) {
	return &dbStorage[T]{
		db:     d.db,
		prefix: fmt.Sprintf("user/%s/", d.name),
	}, nil
}

func (d *dbStorage[T]) dbID(id string) string {
	return d.prefix + id
}

// Delete implements Storage.
func (d *dbStorage[T]) Delete(ctx context.Context, item T) (bool, error) {
	id := d.dbID(item.GetID())
	logging.FromPrefix(ctx, "db-storage").Debug("deleting item: " + id)

	// Deletes by primary key, a string condition would be used as raw SQL
	tx := d.db.WithContext(ctx).Delete(&models.KeyValue{Key: id})

	return tx.RowsAffected > 0, tx.Error
}

// Get implements Storage.
func (d *dbStorage[T]) Get(ctx context.Context, id string) (T, error) {
	id = d.dbID(id)
	logging.FromPrefix(ctx, "db-storage").Debug("reading item: " + id)

	kv := models.KeyValue{
		Key: id,
	}

	tx := d.db.WithContext(ctx).Take(&kv)
	if tx.Error != nil {
		return generics.Empty[T](), tx.Error
	}

	return unmarshallGeneric[T](kv.Value)
}

// Set implements Storage.
func (d *dbStorage[T]) Set(ctx context.Context, item T) error {
	id := d.dbID(item.GetID())
	logging.FromPrefix(ctx, "db-storage").Debug("setting item: " + id)

	data, err := xformat.Marshal(item)
	if err != nil {
		return err
	}

	kv := models.KeyValue{
		Key:   id,
		Value: data,
	}
	cl := clause.OnConflict{
		UpdateAll: true,
	}

	tx := d.db.WithContext(ctx).Clauses(cl).Create(&kv)

	return tx.Error
}

func (d *dbStorage[T]) Keys(ctx context.Context) iter.Seq[string] {
	return func(yield func(string) bool) {
		var kvs []models.KeyValue
		tx := d.db.WithContext(ctx).Where("`key` LIKE ?", d.prefix+"%").Select("key").Find(&kvs)
		if tx.Error != nil {
			return
		}
		for _, kv := range kvs {
			select {
			case <-ctx.Done():
				return
			default:
			}
			dbID, _ := strings.CutPrefix(kv.Key, d.prefix)
			if !yield(dbID) {
				return
			}
		}
	}
}
