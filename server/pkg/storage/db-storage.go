package storage

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"github.com/DaanV2/mechanus/server/internal/logging"
	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/DaanV2/mechanus/server/pkg/database/models"
	xencoding "github.com/DaanV2/mechanus/server/pkg/extensions/encoding"
	"github.com/daanv2/go-kit/generics"

	"gorm.io/gorm/clause"
)

func DBStorage[T Identifiable](db *database.DB) StorageProvider[T] {
	return &dbProvider[T]{db}
}

type dbProvider[T Identifiable] struct {
	db *database.DB
}

type dbStorage[T Identifiable] struct {
	db     *database.DB
	prefix string
}

// AppStorage implements StorageProvider.
func (d *dbProvider[T]) AppStorage() (Storage[T], error) {
	return &dbStorage[T]{
		db:     d.db,
		prefix: fmt.Sprintf("app/%s/", generics.NameOf[T]()),
	}, nil
}

// StateStorage implements StorageProvider.
func (d *dbProvider[T]) StateStorage() (Storage[T], error) {
	return &dbStorage[T]{
		db:     d.db,
		prefix: fmt.Sprintf("state/%s/", generics.NameOf[T]()),
	}, nil
}

// UserStorage implements StorageProvider.
func (d *dbProvider[T]) UserStorage() (Storage[T], error) {
	return &dbStorage[T]{
		db:     d.db,
		prefix: fmt.Sprintf("user/%s/", generics.NameOf[T]()),
	}, nil
}

func (d *dbStorage[T]) dbID(id string) string {
	return d.prefix + id
}

// Delete implements Storage.
func (d *dbStorage[T]) Delete(ctx context.Context, item T) (bool, error) {
	id := d.dbID(item.GetID())
	logging.FromPrefix(ctx, "db-storage").Debug("deleting item: " + id)

	var kv models.KeyValue
	tx := d.db.WithContext(ctx).Limit(1).Delete(&kv, id)

	return tx.RowsAffected > 0, tx.Error
}

// Get implements Storage.
func (d *dbStorage[T]) Get(ctx context.Context, id string) (T, error) {
	id = d.dbID(id)
	logging.FromPrefix(ctx, "db-storage").Debug("reading item: " + id)

	var kv models.KeyValue

	tx := d.db.WithContext(ctx).Take(&kv, id)
	if tx.Error != nil {
		return generics.Empty[T](), tx.Error
	}

	var result T
	err := xencoding.Unmarshal(kv.Value, result)

	return result, err
}

// Set implements Storage.
func (d *dbStorage[T]) Set(ctx context.Context, item T) error {
	id := d.dbID(item.GetID())
	logging.FromPrefix(ctx, "db-storage").Debug("setting item: " + id)

	data, err := xencoding.Marshal(item)
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
