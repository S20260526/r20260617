package app

import (
	"context"
	"internal/app/msgqueue"
)

type JanitorMetrics interface {
	ExportableMetrics

	RegIn()
	RegErr()
}

type Janitor struct {
	Puller  *msgqueue.Puller
	Storage Storage
	Metrics JanitorMetrics
}

func (j *Janitor) Pull(ctx context.Context) error {
	in, err := j.Puller.Pull(ctx)

	if err != nil {
		return err
	}

	j.Metrics.RegIn()

	defer func() {
		if err != nil {
			j.Metrics.RegErr()
		}
	}()

	defer in.Acknowledge()

	out, err := UnmarshalOrder(in.GetData())

	if err == nil {
		err = j.Storage.Delete(ctx, out.BlobId)
	}

	return err
}
