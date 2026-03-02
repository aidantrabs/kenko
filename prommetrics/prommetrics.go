package prommetrics

import (
	"github.com/aidantrabs/kenko"
	"github.com/prometheus/client_golang/prometheus"
)

type Option func(*Reporter)

func WithRegistry(r prometheus.Registerer) Option {
	return func(rep *Reporter) { rep.registerer = r }
}

func WithNamespace(ns string) Option {
	return func(r *Reporter) { r.namespace = ns }
}

type Reporter struct {
	registerer prometheus.Registerer
	namespace  string

	checkDuration *prometheus.HistogramVec
	checkTotal    *prometheus.CounterVec
	targetUp      *prometheus.GaugeVec
}

func New(opts ...Option) *Reporter {
	r := &Reporter{
		registerer: prometheus.DefaultRegisterer,
	}
	for _, opt := range opts {
		opt(r)
	}

	r.checkDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: r.namespace,
		Name:      "kenko_check_duration_seconds",
		Help:      "duration of health checks",
		Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"target"})

	r.checkTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: r.namespace,
		Name:      "kenko_check_total",
		Help:      "total number of health checks",
	}, []string{"target", "status"})

	r.targetUp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: r.namespace,
		Name:      "kenko_target_up",
		Help:      "whether a target is healthy (1) or not (0)",
	}, []string{"target"})

	r.registerer.MustRegister(r.checkDuration, r.checkTotal, r.targetUp)

	return r
}

func (r *Reporter) ReportCheck(target string, status kenko.Status, latencySeconds float64) {
	r.checkDuration.WithLabelValues(target).Observe(latencySeconds)
	r.checkTotal.WithLabelValues(target, string(status)).Inc()

	if status == kenko.StatusHealthy {
		r.targetUp.WithLabelValues(target).Set(1)
	} else {
		r.targetUp.WithLabelValues(target).Set(0)
	}
}
