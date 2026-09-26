package app

import (
	"context"
)

type Storage interface {
	Create(ctx context.Context, blob []byte) (string, error)
	Read(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}
