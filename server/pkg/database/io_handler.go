package database

type IOHandler interface {
	Get(id string) ([]byte, error)
	Set(id string, data []byte) error
}
