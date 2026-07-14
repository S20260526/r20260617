package main

import (
	"log/slog"
	"net/http"
	"os"

	"app"
	"infra"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	s := http.Server{
		Addr:    ":8080",
		Handler: infra.NewHandler(app.NewWorldService()),
	}

	slog.Info("server", "state", "started")

	e := s.ListenAndServe()

	if e != nil {
		panic(e)
	}
}
