// package kenko provides health check monitoring as an importable SDK
// root package has zero third-party dependencies - opt into redis or
// prometheus via the [redisstore] and [prommetrics] sub-packages.
package kenko

import (
	"context"
	"net/http"
)

type Kenko struct {
	checker *Checker
}

func New(opts ...Option) (*Kenko, error) {
	c, err := NewChecker(opts...)
	if err != nil {
		return nil, err
	}
	return &Kenko{checker: c}, nil
}

func (k *Kenko) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/health", HandleHealth(k.checker))
	mux.HandleFunc("/ready", HandleReady(k.checker))
	mux.HandleFunc("/status", HandleStatus(k.checker))
}

func (k *Kenko) Run(ctx context.Context) {
	k.checker.Run(ctx)
}

func (k *Kenko) Checker() *Checker {
	return k.checker
}
