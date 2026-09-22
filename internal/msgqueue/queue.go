package msgqueue

import (
	"context"
)

type Queue interface {
	Connect(url string) error
	OpenChannel() error
	CloseChannel()
	Disconnect()
	Publish(ctx context.Context, msg []byte) error
	Consume(ctx context.Context) ([]byte, error)
}
