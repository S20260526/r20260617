package app

import (
	"context"
	"internal/app/msgqueue"
	//"time"
)

type Back struct {
	Puller      *msgqueue.Puller
	Storage     Storage
	Coprocess   Coprocess
	EventsTable string
	Registrator Registrator
}

func (b *Back) Pull(ctx context.Context) error {
	in, err := b.Puller.Pull(ctx)

	if err != nil {
		return err
	}

	doReject := true

	defer func() {
		if doReject {
			in.Reject()
		} else {
			in.Acknowledge()
		}
	}()

	ord, err := UnmarshalOrder(in.GetData())

	if err != nil {
		return err
	}

	blob, err := b.Storage.Read(ctx, ord.BlobId)

	if err != nil {
		return err
	}

	rsps, err := b.Coprocess.Call(ctx, &CoprocessRequest{Payload: blob})

	if err == nil {
		doReject = false

		switch rsps.Result {
		case CoprocessResultYes:
			err = b.Registrator.Put(
				ctx, b.EventsTable,
				Event{Timestamp: ord.Timestamp, Id: ord.BlobId},
			)

			if err != nil {
				doReject = true
			}

		case CoprocessResultNo:
			b.Storage.Delete(ctx, ord.BlobId)
		default:
			doReject = true
		}
	}

	return err
}
