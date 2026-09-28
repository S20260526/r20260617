package app

import (
	"context"
	"internal/app/msgqueue"
	"time"
)

type Front struct {
	BrokerUrl HostColonPort
	Storage   Storage
	Queue     msgqueue.PublishingQueue
}

func (f *Front) Push(ctx context.Context, t time.Time, blob []byte) error {
	id, err := f.Storage.Create(ctx, blob)

	if err != nil {
		return err
	}

	om, err := Order{Timestamp: t, BlobId: id}.Marshal()

	if err == nil {
		p := msgqueue.NewPusher(f.BrokerUrl.String(), f.Queue)

		defer f.Queue.CloseChannel()

		p.Charge(om)

		err = p.Push(ctx)
	}

	if err != nil {
		f.Storage.Delete(ctx, id)
	}

	return err
}
