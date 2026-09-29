package app

import (
	"context"
	"internal/app/msgqueue"
)

type Janitor struct {
	Puller  *msgqueue.Puller
	Storage Storage
}

func (j *Janitor) Pull(ctx context.Context) error {
	in, err := j.Puller.Pull(ctx)

	if err != nil {
		return err
	}

	defer in.Acknowledge()

	out, err := UnmarshalOrder(in.GetData())

	if err == nil {
		err = j.Storage.Delete(ctx, out.BlobId)
	}

	return err
}
