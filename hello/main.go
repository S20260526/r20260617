package main

import (
	"log/slog"
	"os"

	"app"
	"infra"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	e := infra.StartMTLSServer(infra.NewHandler(app.NewHelloService()))

	slog.Info("hello", "error", e)

	os.Exit(1)
}
