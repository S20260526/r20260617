package app

import (
	"context"
	"internal/app/msgqueue"
	"sync"
	"time"
)

type Front struct {
	Pusher  *msgqueue.Pusher
	Storage Storage
	mutex   sync.Mutex
}

func (f *Front) Push(ctx context.Context, t time.Time, blob []byte) error {
	id, err := f.Storage.Create(ctx, blob)

	if err != nil {
		return err
	}

	om, err := Order{Timestamp: t, BlobId: id}.Marshal()

	if err == nil {
		func() {
			f.mutex.Lock()
			defer f.mutex.Unlock()

			err = f.Pusher.Push(ctx, om)
		}()
	}

	if err != nil {
		f.Storage.Delete(ctx, id)
	}

	return err
}
