package main

import (
	"context"
	"fmt"
	"internal/app"
	"internal/app/msgqueue"
	"internal/infra"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type syncPusher struct {
	Pusher app.Pusher

	mutex sync.Mutex
}

func (p *syncPusher) Push(ctx context.Context, blob []byte) error {
	p.mutex.Lock()

	defer p.mutex.Unlock()

	return p.Pusher.Push(ctx, blob)
}

type handler struct {
	front *app.Front
	mutex sync.Mutex
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" && r.Method == "POST" {
		w.WriteHeader(h.servePostRoot(r))
	}
}

func (h *handler) servePostRoot(r *http.Request) int {
	blob, err := io.ReadAll(r.Body)

	if err != nil {
		return http.StatusBadRequest
	}

	front := func() *app.Front {
		h.mutex.Lock()

		defer h.mutex.Unlock()

		return h.front
	}()

	err = front.Push(r.Context(), time.Now(), blob)

	if err != nil {
		slog.Warn(
			"front",
			"where", "HTTP",
			"when", "serve",
			"what", err,
		)

		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func newFront(cfg app.Configuration) *app.Front {
	return &app.Front{
		Pusher: &syncPusher{
			// TODO what about cleanup?
			Pusher: msgqueue.NewPusher(
				cfg.Pushing.Host.String(),
				infra.NewRMQPublishing(cfg.Pushing.Queue),
			),
		},
		Storage: infra.NewSeaWeedFS(cfg.Storage.String()),
		Metrics: infra.NewPrometrics(),
	}
}

func newServer(cfg app.Configuration, h http.Handler) *http.Server {
	a := fmt.Sprintf(":%d", cfg.InputPort)
	srv := &http.Server{
		Addr:    a,
		Handler: h,
	}

	slog.Info(
		"front",
		"where", "HTTP",
		"when", "starting serve",
		"what", a,
	)

	go func() {
		srv.ListenAndServe()
	}()

	return srv
}

func main() {
	var hndlr *handler
	var srv *http.Server
	var oldCfg app.Configuration

	m := infra.MainObj{
		Tag: "front",
		Setup: func(cfg app.Configuration) error {
			hndlr = &handler{front: newFront(cfg)}
			srv = newServer(cfg, hndlr)
			oldCfg = cfg

			return nil
		},
		Reinit: func(newCfg app.Configuration) error {

			func() {
				hndlr.mutex.Lock()

				defer hndlr.mutex.Unlock()

				oldCfg = newCfg
				hndlr.front = newFront(newCfg)
			}()

			if newCfg.InputPort != oldCfg.InputPort {
				srv.Shutdown(context.Background())

				srv = newServer(newCfg, hndlr)
			}

			return nil
		},
	}

	err := m.MainFunc()

	if err != nil {
		slog.Info("front", "where", "main", "when", "main", "what", err)
	}
}
