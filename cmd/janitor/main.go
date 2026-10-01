package main

import (
	"context"
	"internal/app"
	"internal/app/msgqueue"
	"internal/infra"
	"log/slog"
	"sync"
	"time"
)

var barrier sync.WaitGroup

func newJanitor(ctx context.Context, cfg app.Configuration) *app.Janitor {
	j := app.Janitor{
		Puller: msgqueue.NewPuller(
			app.HostPortUrls(cfg.Pulling.Host, "amqp://"),
			infra.NewRMQConsuming(cfg.Pulling.DLQ),
		),
		Storage: infra.NewSeaWeedFS(cfg.Storage.String()),
		Metrics: infra.NewPrometrics(),
	}

	barrier.Add(1)

	go func() {
		defer barrier.Done()

		for ctx.Err() == nil {
			err := j.Pull(ctx)

			if err != nil {
				slog.Warn(
					"janitor",
					"where", "Pull",
					"when", "Pull",
					"what", err,
				)

				select {
				case <-time.After(cfg.Pulling.RetryTO):
				case <-ctx.Done():
				}
			}
		}

		j.Puller.Cleanup()

		slog.Info(
			"janitor",
			"where", "Cleanup",
			"when", "Cleanup",
			"what", "completed",
		)
	}()

	return &j
}

func main() {
	var ctx, cancel = context.WithCancel(context.Background())

	mainObj := infra.MainObj{Tag: "janitor"}
	mainObj.Setup = func(cfg app.Configuration) error {
		jntr := newJanitor(ctx, cfg)

		mainObj.ExportMetrics(jntr.Metrics)

		return nil
	}
	mainObj.Reinit = func(cfg app.Configuration) error {
		cancel()

		ctx, cancel = context.WithCancel(context.Background())

		jntr := newJanitor(ctx, cfg)

		mainObj.ExportMetrics(jntr.Metrics)

		return nil
	}

	mainObj.MainFunc()

	cancel()

	barrier.Wait()
}
