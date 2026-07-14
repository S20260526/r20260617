package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type EntryHandler struct {
	requestCount, hcheckCount       prometheus.Counter
	promHttpHandler, requestHandler http.Handler
}

func relay(tag, url string) chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)

		cl := http.Client{
			Timeout: 5 * time.Second,
		}

		r, e := cl.Get(url)

		if e != nil {
			slog.Info(tag, "error", e)
			return
		}

		defer r.Body.Close()

		if r.StatusCode != 200 {
			slog.Info(tag, "status", r.StatusCode)
			return
		}

		b, e := io.ReadAll(r.Body)

		if e != nil {
			slog.Info(tag, "read error", e)
			return
		}

		ch <- string(b)
	}()

	return ch
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	helloCh := relay("hello", "http://m20260618-hello")
	worldCh := relay("world", "http://m20260618-world")

	hello := <-helloCh
	world := <-worldCh

	if hello == "" || world == "" {
		w.WriteHeader(500)
		return
	}

	w.Write([]byte(fmt.Sprintf("%s, %s!", hello, world)))
}

func (h EntryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/hcheck" {
		h.hcheckCount.Inc()
	} else if r.RequestURI == "/prometrics" {
		h.promHttpHandler.ServeHTTP(w, r)
	} else {
		h.requestCount.Inc()

		requestHandler(w, r)
	}
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	pr := prometheus.NewRegistry()

	pr.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	h := EntryHandler{
		requestCount: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "request_cnt",
				Help: "Hello, world! request counter",
			},
		),
		hcheckCount: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "hcheck_cnt",
				Help: "Healthcheck request counter",
			},
		),
		promHttpHandler: promhttp.HandlerFor(pr, promhttp.HandlerOpts{Registry: pr}),
	}

	pr.MustRegister(h.requestCount)
	pr.MustRegister(h.hcheckCount)

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
