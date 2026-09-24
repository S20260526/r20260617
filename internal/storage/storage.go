package storage

type Storage interface {
	Create(blob []byte) (string, error)
	Read(key string) ([]byte, error)
	Delete(key string) error
}
