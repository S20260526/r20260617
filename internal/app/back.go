package app

import (
	"context"
	"internal/app/msgqueue"
	//"time"
)

type Back struct {
	Puller    *msgqueue.Puller
	Storage   Storage
	Coprocess Coprocess
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
		switch rsps.Result {
		case CoprocessResultNo:
			b.Storage.Delete(ctx, ord.BlobId)
			fallthrough
		case CoprocessResultYes:
			doReject = false
		}
	}

	return err
}
