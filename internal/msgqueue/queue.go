package msgqueue

import (
	"context"
)

type Incoming interface {
	GetData() []byte
	Acknowledge()
	Reject()
}

type Queue interface {
	Connect(url string) error
	OpenChannel() error
	CloseChannel()
	Disconnect()
	Publish(ctx context.Context, msg []byte) error
	Consume(ctx context.Context) (Incoming, error)
}
