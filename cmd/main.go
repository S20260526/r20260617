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
	pc prometheus.Counter
	ih infra.Handler
	ph http.Handler
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/hcheck" {
		h.pc.Inc()
		return
	} else if r.RequestURI == "/prometrics" {
		h.ph.ServeHTTP(w, r)
	} else {
		h.ih.ServeHTTP(w, r)
	}
}

type Metrics struct {
	registry *prometheus.Registry
}

func (m Metrics) NewCounter(name, human_readable string) app.Counter {
	c := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: name,
			Help: human_readable,
		},
	)

	m.registry.MustRegister(c)

	return c
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	pr := prometheus.NewRegistry()

	pr.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	m := Metrics{
		registry: pr,
	}

	h := Handler{
		pc: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "hcheck_cnt",
				Help: "Healthcheck request counter",
			},
		),
		ih: infra.NewHandler(app.NewService(m)),
		ph: promhttp.HandlerFor(pr, promhttp.HandlerOpts{Registry: pr}),
	}

	pr.MustRegister(h.pc)

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
