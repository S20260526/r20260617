package main

import (
	"app"
	"infra"
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler struct {
	ih infra.Handler
	ph http.Handler
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/hcheck" {
		return
	} else if r.RequestURI == "/prometrics" {
		h.ph.ServeHTTP(w, r)
	} else {
		h.ih.ServeHTTP(w, r)
	}
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	pr := prometheus.NewRegistry()

	pr.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	h := Handler{
		ih: infra.NewHandler(app.NewService()),
		ph: promhttp.HandlerFor(pr, promhttp.HandlerOpts{Registry: pr}),
	}

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
