package app

import (
	"context"
	"internal/app/msgqueue"
)

type Worker struct {
	Puller      *msgqueue.Puller
	Storage     Storage
	Coprocess   Coprocess
	EventsTable string
	Registrator Registrator
}

func (w *Worker) Pull(ctx context.Context) error {
	in, err := w.Puller.Pull(ctx)

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

	blob, err := w.Storage.Read(ctx, ord.BlobId)

	if err != nil {
		return err
	}

	rsps, err := w.Coprocess.Call(ctx, &CoprocessRequest{Payload: blob})

	if err == nil {
		doReject = false

		switch rsps.Result {
		case CoprocessResultYes:
			err = w.Registrator.Put(
				ctx, w.EventsTable,
				Event{Timestamp: ord.Timestamp, Id: ord.BlobId},
			)

			if err != nil {
				doReject = true
			}

		case CoprocessResultNo:
			w.Storage.Delete(ctx, ord.BlobId)
		default:
			doReject = true
		}
	}

	return err
}
