// package kenko provides health check monitoring as an importable SDK
// root package has zero third-party dependencies - opt into redis or
// prometheus via the [redisstore] and [prommetrics] sub-packages.
package kenko

import (
	"context"
	"net/http"
)

// Kenko is the top-level entry point that wires together a health checker and HTTP handlers.
type Kenko struct {
	checker *Checker
}

// New creates a Kenko instance configured with the given options.
func New(opts ...Option) (*Kenko, error) {
	c, err := NewChecker(opts...)
	if err != nil {
		return nil, err
	}
	return &Kenko{checker: c}, nil
}

// RegisterHandlers registers the /health, /ready, and /status HTTP handlers on the given mux.
func (k *Kenko) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/health", HandleHealth(k.checker))
	mux.HandleFunc("/ready", HandleReady(k.checker))
	mux.HandleFunc("/status", HandleStatus(k.checker))
}

// Run starts the periodic health check loop, blocking until ctx is cancelled.
func (k *Kenko) Run(ctx context.Context) {
	k.checker.Run(ctx)
}

// Checker returns the underlying Checker for direct access.
func (k *Kenko) Checker() *Checker {
	return k.checker
}
