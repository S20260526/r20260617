package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"infra"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type EntryHandler struct {
	transport                 *http.Transport
	requestCount, hcheckCount prometheus.Counter
	promHttpHandler           http.Handler
}

func (h *EntryHandler) relay(tag, url string) chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)

		cl := http.Client{
			Transport: h.transport,
			Timeout:   5 * time.Second,
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

func (h *EntryHandler) requestHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("request", "from", r.RemoteAddr, "to", r.Host, "URI", r.RequestURI)

	h.requestCount.Inc()

	helloCh := h.relay("hello", "https://hello:8081")
	worldCh := h.relay("world", "https://world:8082")

	hello := <-helloCh
	world := <-worldCh

	if hello == "" || world == "" {
		w.WriteHeader(500)
		return
	}

	w.Write([]byte(fmt.Sprintf("%s, %s!", hello, world)))
}

func (h *EntryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/hcheck" {
		h.hcheckCount.Inc()
	} else if r.RequestURI == "/prometrics" {
		h.promHttpHandler.ServeHTTP(w, r)
	} else {
		h.requestHandler(w, r)
	}
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	tc := &tls.Config{}

	var e error

	tc.Certificates, e = infra.ServerCertFromFile("client.crt", "client.key")

	if e != nil {
		panic(e)
	}

	tc.ClientCAs, e = infra.CaCertFromFile("ca.crt")

	if e != nil {
		panic(e)
	}

	tc.RootCAs, e = infra.CaCertFromFile("ca.crt")

	if e != nil {
		panic(e)
	}

	pr := prometheus.NewRegistry()

	pr.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	h := EntryHandler{
		transport: &http.Transport{
			TLSClientConfig: tc,
		},
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
		Handler: &h,
	}

	slog.Info("server", "state", "started")

	e = s.ListenAndServe()

	if e != nil {
		panic(e)
	}
}
