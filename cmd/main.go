package main

import (
	"app"
	"infra"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	h := infra.NewHandler(app.NewService())

	s := http.Server{
		Addr:    ":8080",
		Handler: h,
	}

	slog.Info("server", "state", "started")

	e := s.ListenAndServe()

	if e != nil {
		panic(e)
	}
}
