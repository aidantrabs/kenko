package kenko

import (
	"log/slog"
	"net/http"
	"time"
)

// Option configures a Kenko instance.
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

// WithTarget adds a named URL to the list of endpoints to check.
func WithTarget(name, url string) Option {
	return func(o *options) {
		o.targets = append(o.targets, Target{Name: name, URL: url})
	}
}

// WithInterval sets the duration between check cycles (default 30s).
func WithInterval(d time.Duration) Option {
	return func(o *options) { o.interval = d }
}

// WithTimeout sets the HTTP request timeout per check (default 5s).
func WithTimeout(d time.Duration) Option {
	return func(o *options) { o.timeout = d }
}

// WithStore sets the result store (default MemoryStore).
func WithStore(s Store) Option {
	return func(o *options) { o.store = s }
}

// WithMetrics sets the MetricsReporter used to record check results.
func WithMetrics(m MetricsReporter) Option {
	return func(o *options) { o.metrics = m }
}

// WithLogger sets the structured logger (default slog.Default).
func WithLogger(l *slog.Logger) Option {
	return func(o *options) { o.logger = l }
}

// WithHTTPClient sets a custom HTTP client for health checks.
func WithHTTPClient(c *http.Client) Option {
	return func(o *options) { o.client = c }
}
