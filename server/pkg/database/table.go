package database

import (
	"encoding"
	"encoding/json"
	"iter"
	"path/filepath"

	"github.com/charmbracelet/log"
)

type TableName string

var _ IOHandler = &rawTable{}

type Table[T any] struct {
	table      *rawTable
	logger *log.Logger
}

type rawTable struct {
	name    string
	handler IOHandler
}

// Ids implements IOHandler.
func (r *rawTable) Ids() iter.Seq[string] {
	return r.handler.Ids()
}

// Get implements IOHandler.
func (r *rawTable) Get(id string) ([]byte, error) {
	return r.handler.Get(filepath.Join(r.name, id))
}

// Set implements IOHandler.
func (r *rawTable) Set(id string, data []byte) error {
	return r.handler.Set(filepath.Join(r.name, id), data)
}

func (r *rawTable) Name() string {
	return r.name
}

func newTable[T any](name TableName, handler IOHandler) *Table[T] {
	return &Table[T]{
		table: &rawTable{
			name:    string(name),
			handler: handler,
		},
		logger: log.Default().WithPrefix("db[\"" + string(name) + "\"]"),
	}
}

func (r *Table[T]) Ids() iter.Seq[string] {
	return r.table.Ids()
}

func (table *Table[T]) Name() string {
	return table.table.name
}

func (table *Table[T]) String() string {
	return "db.table[\"" + table.table.name + "\"]"
}

func (table *Table[T]) Get(id string) (T, error) {
	table.logger.Debugf("retrieving '%s' from db '%s'", id, table.table.name)

	var result T
	data, err := table.table.Get(id)
	if err != nil {
		return result, err
	}

	if v, ok := interface{}(result).(encoding.BinaryUnmarshaler); ok {
		err = v.UnmarshalBinary(data)
	} else {
		err = json.Unmarshal(data, &result)
	}

	return result, err
}

func (table *Table[T]) Set(id string, item T) error {
	table.logger.Debugf("setting '%s' to db '%s'", id, table.table.name)
	var (
		data []byte
		err  error
	)

	if v, ok := interface{}(item).(encoding.BinaryMarshaler); ok {
		data, err = v.MarshalBinary()
	} else {
		data, err = json.Marshal(item)
	}
	if err != nil {
		return err
	}

	return table.table.Set(id, data)
}

// First returns the item that matches the given predicate first, returns [ErrNotFound] is nothing is found
func (table *Table[T]) First(predicate func(item T) bool) (T, error) {
	var empty T
	for id := range table.Ids() {
		v, err := table.Get(id)
		if err != nil {
			return empty, err
		}
		if predicate(v) {
			return v, nil
		}
	}

	return empty, ErrNotFound
}