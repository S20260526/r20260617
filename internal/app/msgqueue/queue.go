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
}

type PublishingQueue interface {
	Queue
	Publish(ctx context.Context, msg []byte) error
}

type ConsumingQueue interface {
	Queue
	Consume(ctx context.Context) (Incoming, error)
}
