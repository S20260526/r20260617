package infra

import (
	"context"
	"errors"
	"fmt"
	"internal/app"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type MainObj struct {
	Tag string

	Setup, Reinit func(app.Configuration) error

	etcd    *Etcd
	sigChan <-chan os.Signal
	cfgChan <-chan app.Configuration
	cfg     app.Configuration

	metricsSrv *http.Server
}

func (m *MainObj) newEtcd() (*Etcd, error) {
	var url = "http://localhost:2379"

	if len(os.Args) >= 2 {
		url = os.Args[1]
	}

	slog.Info(
		m.Tag,
		"where", "etcd",
		"when", "args",
		"what", url,
	)

	return NewEtcd(url)
}

func (m *MainObj) newSigChan() <-chan os.Signal {
	chn := make(chan os.Signal, 0)

	signal.Notify(chn, syscall.SIGINT, syscall.SIGTERM)

	return chn
}

func (m *MainObj) MainFunc() {
	err := m.mainFunc()

	if err != nil {
		slog.Info(m.Tag, "where", "main", "when", "main", "what", err)
	}
}

func (m *MainObj) ExportMetrics(metrics app.ExportableMetrics) {
	if m.metricsSrv != nil {
		slog.Warn(
			m.Tag,
			"where", "metrics",
			"when", "export",
			"what", "server already running",
		)
	}

	m.metricsSrv = &http.Server{
		Addr:    fmt.Sprintf(":%d", m.cfg.MetricsPort),
		Handler: metrics.GetHttpHandler(),
	}

	go func() {
		err := m.metricsSrv.ListenAndServe()

		if err != nil {
			slog.Info(
				m.Tag,
				"where", "metrics",
				"when", "export",
				"what", err,
			)
		}
	}()
}

func (m *MainObj) mainFunc() error {
	var err error

	slog.Info(m.Tag, "where", "mainfunc", "when", "config", "what", "initing")

	m.etcd, err = m.newEtcd()

	if err != nil {
		return err
	}

	slog.Info(m.Tag, "where", "mainfunc", "when", "signotify", "what", "starting")

	m.sigChan = m.newSigChan()

	slog.Info(m.Tag, "where", "mainonbj", "when", "config", "what", "starting")

	etcdCtx, etcdCancel := context.WithCancel(context.Background())

	defer etcdCancel()

	m.cfg, m.cfgChan, err = m.etcd.Watch(etcdCtx)

	if err != nil {
		return err
	}

	slog.Info(m.Tag, "where", "mainonbj", "when", "loop", "what", "starting")

	defer func() {
		if m.metricsSrv != nil {
			m.metricsSrv.Shutdown(context.Background())
		}
	}()

	if m.Setup != nil {
		err = m.Setup(m.cfg)
	}

	for err == nil {
		err = m.selectCfgSig()
	}

	slog.Info(m.Tag, "where", "mainonbj", "when", "loop", "what", "exit")

	return err
}

func (m *MainObj) selectCfgSig() error {
	select {
	case _ = <-m.sigChan:
		return errors.New("exit on signal")
	case cfg, ok := <-m.cfgChan:
		if !ok {
			return errors.New("configuration watcher channel broken")
		}

		m.cfg = cfg

		if m.metricsSrv != nil {
			m.metricsSrv.Shutdown(context.Background())
			m.metricsSrv = nil
		}

		if m.Reinit != nil {
			m.Reinit(cfg)
		}
	}

	return nil
}
