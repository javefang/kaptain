package store

type Store interface {
	List(key string) ([]string, error)
	Exists(key string) (bool, error)
	Get(key string) ([]byte, error)
	Set(key string, data []byte) error
	Delete(key string) error
	DeleteAll(key string) error
}
