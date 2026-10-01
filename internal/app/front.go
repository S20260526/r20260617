package app

import (
	"context"
	"net/http"
	"time"
)

type FrontMetrics interface {
	RegIn()
	RegOut()
	RegErr()
	GetHttpHandler() http.Handler
}

type Pusher interface {
	Push(ctx context.Context, blob []byte) error
}

type Front struct {
	Pusher  Pusher
	Storage Storage
	Metrics FrontMetrics
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
		err = f.Pusher.Push(ctx, om)
	}

	if err != nil {
		f.Storage.Delete(ctx, id)
	}

	return err
}
