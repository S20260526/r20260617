package main

import (
	"context"
	"errors"
	"fmt"
	"internal/app"
	"internal/app/msgqueue"
	"internal/infra"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

func newEtcd() (*infra.Etcd, error) {
	var url = "http://localhost:2379"

	if len(os.Args) >= 2 {
		url = os.Args[1]
	}

	slog.Info(
		"front",
		"where", "etcd",
		"when", "args",
		"what", url,
	)

	return infra.NewEtcd(url)
}

func newSigChan() <-chan os.Signal {
	chn := make(chan os.Signal, 0)

	signal.Notify(chn, syscall.SIGINT, syscall.SIGTERM)

	return chn
}

func newFront(cfg app.Configuration) *app.Front {
	return &app.Front{
		Pusher: &syncPusher{
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

type mainObj struct {
	etcd    *infra.Etcd
	sigChan <-chan os.Signal
	cfgChan <-chan app.Configuration

	oldCfg app.Configuration

	hndlr *handler
	srv   *http.Server
}

func (m *mainObj) mainFunc() error {
	var err error

	slog.Info("front", "where", "mainonbj", "when", "config", "what", "initing")

	m.etcd, err = newEtcd()

	if err != nil {
		return err
	}

	slog.Info("front", "where", "mainonbj", "when", "signotify", "what", "starting")

	m.sigChan = newSigChan()

	slog.Info("front", "where", "mainonbj", "when", "config", "what", "starting")

	etcdCtx, etcdCancel := context.WithCancel(context.Background())

	defer etcdCancel()

	m.oldCfg, m.cfgChan, err = m.etcd.Watch(etcdCtx)

	if err != nil {
		return err
	}

	slog.Info("front", "where", "mainonbj", "when", "loop", "what", "starting")

	m.hndlr = &handler{front: newFront(m.oldCfg)}
	m.srv = newServer(m.oldCfg, m.hndlr)

	for err == nil {
		err = m.selectCfgSig()
	}

	slog.Info("front", "where", "mainonbj", "when", "loop", "what", "exit")

	return err
}

func (m *mainObj) selectCfgSig() error {
	select {
	case _ = <-m.sigChan:
		return errors.New("exit on signal")
	case newCfg, ok := <-m.cfgChan:
		if !ok {
			return errors.New("configuration watcher channel broken")
		}

		func() {
			m.hndlr.mutex.Lock()

			defer m.hndlr.mutex.Unlock()

			m.oldCfg = newCfg
			m.hndlr.front = newFront(newCfg)
		}()

		if newCfg.InputPort != m.oldCfg.InputPort {
			m.srv.Shutdown(context.Background())

			m.srv = newServer(newCfg, m.hndlr)
		}
	}

	return nil
}

func main() {
	m := mainObj{}

	err := m.mainFunc()

	if err != nil {
		slog.Info(
			"front",
			"where", "main",
			"when", "main loop",
			"what", err,
		)
	}
}
