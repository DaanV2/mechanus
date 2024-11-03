package database

import "iter"

type IOHandler interface {
	Get(id string) ([]byte, error)
	Set(id string, data []byte) error
	Ids() iter.Seq[string]
}
