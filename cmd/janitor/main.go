package main

import (
	"context"
	"internal/app"
	"internal/app/msgqueue"
	"internal/infra"
	"log/slog"
)

func newJanitor(ctx context.Context, cfg app.Configuration) {
	j := app.Janitor{
		Puller: msgqueue.NewPuller(
			app.HostPortUrls(cfg.Pulling.Host, "amqp://"),
			infra.NewRMQConsuming(cfg.Pulling.DLQ),
		),
		Storage: infra.NewSeaWeedFS(cfg.Storage.String()),
		Metrics: infra.NewPrometrics(),
	}

	go func() {
		for ctx.Err() == nil {
			err := j.Pull(ctx)

			if err != nil {
				slog.Warn(
					"janitor",
					"where", "Pull",
					"when", "Pull",
					"what", err,
				)
			}
		}
	}()
}

func main() {
	var ctx, cancel = context.WithCancel(context.Background())

	m := infra.MainObj{
		Tag: "janitor",
		Setup: func(cfg app.Configuration) error {
			newJanitor(ctx, cfg)

			return nil
		},
		Reinit: func(cfg app.Configuration) error {
			cancel()

			ctx, cancel = context.WithCancel(context.Background())

			newJanitor(ctx, cfg)

			return nil
		},
	}

	m.MainFunc()

	cancel()
}
