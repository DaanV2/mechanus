package database

import (
	"encoding"
	"encoding/json"
	"path/filepath"

	"github.com/charmbracelet/log"
)

type TableName string

var _ IOHandler = &rawTable{}

type Table[T any] struct {
	t *rawTable
	logger *log.Logger
}

type rawTable struct {
	name    string
	handler IOHandler
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
		t: &rawTable{
			name:    string(name),
			handler: handler,
		},
		logger: log.Default().WithPrefix("db[\""+string(name)+"\"]"),
	}
}

func (table *Table[T]) Name() string {
	return table.t.name
}

func (table *Table[T]) String() string {
	return "db.table[\"" + table.t.name + "\"]"
}

func (table *Table[T]) Get(id string) (T, error) {
	table.logger.Debugf("retrieving '%s' from db '%s'", id, table.t.name)

	var result T
	data, err := table.t.Get(id)
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
	table.logger.Debugf("setting '%s' to db '%s'", id, table.t.name)
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

	return table.t.Set(id, data)
}
