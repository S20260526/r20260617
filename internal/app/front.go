package app

import (
	"context"
	"internal/app/msgqueue"
	"sync"
	"time"
)

type FrontMetrics interface {
	RegIn()
	RegOut()
	RegErr()
}

type Front struct {
	Pusher  *msgqueue.Pusher
	Storage Storage
	Metrics FrontMetrics
	mutex   sync.Mutex
}

func (f *Front) Push(ctx context.Context, t time.Time, blob []byte) error {
	f.Metrics.RegIn()

	id, err := f.Storage.Create(ctx, blob)

	defer func() {
		if err != nil {
			f.Metrics.RegErr()
		} else {
			f.Metrics.RegOut()
		}
	}()

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
