package kenko

import (
	"log/slog"
	"net/http"
	"time"
)

type Option func(*options)

type options struct {
	targets  []Target
	interval time.Duration
	timeout  time.Duration
	store    Store
	metrics  MetricsReporter
	logger   *slog.Logger
	client   *http.Client
}

func defaults() *options {
	return &options{
		interval: 30 * time.Second,
		timeout:  5 * time.Second,
		logger:   slog.Default(),
	}
}

func WithTarget(name, url string) Option {
	return func(o *options) {
		o.targets = append(o.targets, Target{Name: name, URL: url})
	}
}

func WithInterval(d time.Duration) Option {
	return func(o *options) { o.interval = d }
}

func WithTimeout(d time.Duration) Option {
	return func(o *options) { o.timeout = d }
}

func WithStore(s Store) Option {
	return func(o *options) { o.store = s }
}

func WithMetrics(m MetricsReporter) Option {
	return func(o *options) { o.metrics = m }
}

func WithLogger(l *slog.Logger) Option {
	return func(o *options) { o.logger = l }
}

func WithHTTPClient(c *http.Client) Option {
	return func(o *options) { o.client = c }
}
