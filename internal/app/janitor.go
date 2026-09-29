package app

import (
	"context"
	"internal/app/msgqueue"
)

type JanitorMetrics interface {
	RegIn()
	RegErr()
}

type Janitor struct {
	Puller  *msgqueue.Puller
	Storage Storage
	Metrics JanitorMetrics
}

func (j *Janitor) Pull(ctx context.Context) error {
	j.Metrics.RegIn()

	in, err := j.Puller.Pull(ctx)

	defer func() {
		if err != nil {
			j.Metrics.RegErr()
		}
	}()

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
