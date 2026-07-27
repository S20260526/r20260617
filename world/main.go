package main

import (
	"log/slog"
	"os"

	"app"
	"infra"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	e := infra.StartMTLSServer(infra.NewHandler(app.NewWorldService()))

	slog.Info("world", "error", e)

	os.Exit(1)
}
