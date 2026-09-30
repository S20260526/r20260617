package main

import (
	"context"
	"internal/app"
	"internal/app/msgqueue"
	"internal/infra"
	"log/slog"
	"os"
)

func hostPortStrings(in []app.HostColonPort) []string {
	var out []string

	for _, it := range in {
		out = append(out, it.String())
	}

	return out
}

type ipc struct {
	filename string
	ipc      *infra.GrpcIpcClient
}

func newipc() (*ipc, error) {
	f, err := os.CreateTemp(os.TempDir(), "grpcipc.*")

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

func newWorker(ctx context.Context, cfg app.Configuration) error {
	registrator, err := infra.NewRDBMS(
		ctx,
		cfg.Registrator.Driver, cfg.Registrator.Dsn,
	)

	if err != nil {
		return err
	}

	ipc, err := newipc()

	if err != nil {
		return err
	}

	coprocess, err := infra.NewPython3(ctx, cfg.ScriptFile, ipc)

	if err != nil {
		return err
	}

	w := app.Worker{
		Puller: msgqueue.NewPuller(
			hostPortStrings(cfg.Pulling.Host),
			infra.NewRMQConsuming(cfg.Pulling.Queue),
		),
		Storage:     infra.NewSeaWeedFS(cfg.Storage.String()),
		Coprocess:   coprocess,
		EventsTable: "event",
		Registrator: registrator,
		Metrics:     infra.NewPrometrics(),
	}

	go func() {
		for ctx.Err() != nil {
			err := w.Pull(ctx)

			if err != nil {
				slog.Warn(
					"worker",
					"where", "Pull",
					"when", "Pull",
					"what", err,
				)
			}
		}
	}()

	return nil
}

func main() {
	var ctx, cancel = context.WithCancel(context.Background())

	m := infra.MainObj{
		Tag: "worker",
		Setup: func(cfg app.Configuration) error {
			return newWorker(ctx, cfg)
		},
		Reinit: func(cfg app.Configuration) error {
			cancel()

			ctx, cancel = context.WithCancel(context.Background())

			return newWorker(ctx, cfg)
		},
	}

	m.MainFunc()

	cancel()
}
