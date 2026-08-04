package main

import (
	"log/slog"
	"os"
	"sync"

	"app"
	"infra"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	wg := sync.WaitGroup{}

	wg.Add(2)

	go func() {
		e := infra.StartGrpcServer(infra.NewHandler(app.NewWorldService()))

		slog.Info("gRPC", "error", e)

		os.Exit(1)
	}()

	go func() {
		e := infra.StartMTLSServer(infra.NewHandler(app.NewWorldService()))

		slog.Info("HTTP", "error", e)

		os.Exit(1)
	}()

	wg.Wait()

	os.Exit(1)
}
