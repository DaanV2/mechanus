package database

import (
	"sync"

	"github.com/DaanV2/mechanus/server/pkg/config"
	"github.com/charmbracelet/log"
)

type Database struct {
	tables  sync.Map
	handler IOHandler
}

func GetTable[T any](db *Database, name TableName) *Table[T] {
	t, ok := db.tables.Load(name)
	if ok {
		if v, ok := t.(*Table[T]); ok {
			return v
		}
	}

	nt := newTable[T](name, db.handler)
	db.tables.Store(name, nt)

	return nt
}

func NewDatabase() (*Database, error) {
	dir := config.Database.Folder.Value()
	handler, err := NewFileIO(dir)
	if err != nil {
		return nil, err
	}

	log.Info("opening new database", "folder", dir)

	return NewDatabaseWith(handler), nil
}

func NewDatabaseWith(handler IOHandler) *Database {
	return &Database{
		tables:  sync.Map{},
		handler: handler,
	}
}
