package main

import (
	"context"
	"internal/app"
	"internal/app/msgqueue"
	"internal/infra"
	"log/slog"
	"os"
	"sync"
	"time"
)

var barrier sync.WaitGroup

type ipc struct {
	filename string
	ipc      *infra.GrpcIpcClient
}

func newIpc(cfg app.Configuration) (*ipc, error) {
	f, err := os.CreateTemp(cfg.Scripting.SocketDir, "grpcipc.*")

	if err != nil {
		return nil, err
	}

	defer f.Close()

	return &ipc{f.Name(), infra.NewGrpcIpcClient(f.Name())}, nil
}

func (i *ipc) SocketName() string {
	return i.filename
}

func (i *ipc) Call(ctx context.Context, rqst *app.CoprocessRequest) (*app.CoprocessResponse, error) {
	rsps, err := i.ipc.Process(ctx, rqst)

	if err != nil {
		return nil, err
	}

	return rsps, nil
}

func newWorker(ctx context.Context, cfg app.Configuration) (*app.Worker, error) {
	registrator, err := infra.NewRDBMS(
		ctx,
		cfg.Registrator.Driver, cfg.Registrator.Dsn,
	)

	if err != nil {
		return nil, err
	}

	ipc, err := newIpc(cfg)

	if err != nil {
		return nil, err
	}

	coprocess, err := infra.NewPython3(ctx, cfg.Scripting.File, ipc)

	if err != nil {
		return nil, err
	}

	w := app.Worker{
		Puller: msgqueue.NewPuller(
			app.HostPortUrls(cfg.Pulling.Host, "amqp://"),
			infra.NewRMQConsuming(cfg.Pulling.Queue),
		),
		Storage:     infra.NewSeaWeedFS(cfg.Storage.String()),
		Coprocess:   coprocess,
		EventsTable: cfg.Registrator.Table,
		Registrator: registrator,
		Metrics:     infra.NewPrometrics(),
	}

	barrier.Add(1)

	go func() {
		defer barrier.Done()

		for ctx.Err() == nil {
			err := w.Pull(ctx)

			if err != nil {
				slog.Warn(
					"worker",
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

		w.Puller.Cleanup()
		w.Coprocess.Wait()
		w.Registrator.Cleanup()

		slog.Info(
			"worker",
			"where", "Cleanup",
			"when", "Cleanup",
			"what", "completed",
		)
	}()

	return &w, nil
}

func main() {
	var ctx, cancel = context.WithCancel(context.Background())

	mainObj := infra.MainObj{Tag: "worker"}
	mainObj.Setup = func(cfg app.Configuration) error {
		wrk, err := newWorker(ctx, cfg)

		if err == nil {
			mainObj.ExportMetrics(wrk.Metrics)
		} else {
			slog.Warn(
				"worker",
				"where", "Setup",
				"when", "Setup",
				"what", err,
			)
		}

		return nil
	}
	mainObj.Reinit = func(cfg app.Configuration) error {
		cancel()

		ctx, cancel = context.WithCancel(context.Background())
		wrk, err := newWorker(ctx, cfg)

		if err == nil {
			mainObj.ExportMetrics(wrk.Metrics)
		} else {
			slog.Warn(
				"worker",
				"where", "Setup",
				"when", "Setup",
				"what", err,
			)
		}

		return nil
	}

	mainObj.MainFunc()

	cancel()

	barrier.Wait()
}
