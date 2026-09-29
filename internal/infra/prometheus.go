package infra

import (
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	nethttp "net/http"
	"time"
)

const (
	metricIn int = iota
	metricOut
	metricErr
	metricYes
	metricNo
	metricFail

	metricEnd
)

type Prometrics struct {
	registry *prom.Registry
	metrics  [metricEnd]prom.Counter
	duration prom.Histogram
}

func NewPrometrics() *Prometrics {
	m := &Prometrics{registry: prom.NewRegistry()}

	for i, s := range []struct{ name, help string }{
		{"in_cnt", "incoming count"},
		{"out_cnt", "outgoing count"},
		{"err_cnt", "error count"},
		{"yes_cnt", "positive detect count"},
		{"no_cnt", "negative detect count"},
		{"fail_cnt", "failed detect count"},
	} {
		m.metrics[i] = prom.NewCounter(
			prom.CounterOpts{Name: s.name, Help: s.help},
		)

		m.registry.MustRegister(m.metrics[i])
	}

	m.duration = prom.NewHistogram(
		prom.HistogramOpts{
			Name:    "dur",
			Help:    "request duration",
			Buckets: prom.ExponentialBuckets(0.01, 2, 10),
		},
	)

	m.registry.MustRegister(m.duration)

	return m
}

func (m *Prometrics) RegIn() {
	m.metrics[metricIn].Inc()
}

func (m *Prometrics) RegOut() {
	m.metrics[metricOut].Inc()
}

func (m *Prometrics) RegYes() {
	m.metrics[metricYes].Inc()
}

func (m *Prometrics) RegNo() {
	m.metrics[metricNo].Inc()
}

func (m *Prometrics) RegFail() {
	m.metrics[metricFail].Inc()
}

func (m *Prometrics) RegErr() {
	m.metrics[metricErr].Inc()
}

func (m *Prometrics) RegFrontToEndDuration(beginTime time.Time) {
	m.duration.Observe(time.Since(beginTime).Seconds())
}

func (m *Prometrics) GetHttpHandler() nethttp.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}
