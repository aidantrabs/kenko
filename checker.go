package kenko

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsReporter is implemented by types that record health check metrics.
type MetricsReporter interface {
	ReportCheck(target string, status Status, latencySeconds float64)
}

// Checker performs periodic HTTP health checks against configured targets.
type Checker struct {
	client   *http.Client
	store    Store
	targets  []Target
	interval time.Duration
	logger   *slog.Logger
	metrics  MetricsReporter

	ready atomic.Bool
}

// NewChecker creates a Checker configured with the given options.
func NewChecker(opts ...Option) (*Checker, error) {
	o := defaults()
	for _, opt := range opts {
		opt(o)
	}

	if len(o.targets) == 0 {
		return nil, fmt.Errorf("kenko: at least one target is required")
	}

	if o.store == nil {
		o.store = NewMemoryStore()
	}

	client := o.client
	if client == nil {
		client = &http.Client{}
	}
	client.Timeout = o.timeout

	return &Checker{
		client:   client,
		store:    o.store,
		targets:  o.targets,
		interval: o.interval,
		logger:   o.logger,
		metrics:  o.metrics,
	}, nil
}

// Ready reports whether the checker has completed at least one check cycle.
func (c *Checker) Ready() bool { return c.ready.Load() }

// Store returns the result store used by the checker.
func (c *Checker) Store() Store { return c.store }

// Results returns the latest check results for all targets.
func (c *Checker) Results() (map[string]Result, error) {
	return c.store.GetAll(context.Background())
}

// Run starts the check loop, blocking until ctx is cancelled.
func (c *Checker) Run(ctx context.Context) {
	c.logger.Info("checker starting", "targets", len(c.targets), "interval", c.interval)

	c.checkAll(ctx)
	c.ready.Store(true)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("checker stopping")
			return
		case <-ticker.C:
			c.checkAll(ctx)
		}
	}
}

func (c *Checker) checkAll(ctx context.Context) {
	var wg sync.WaitGroup

	for _, target := range c.targets {
		wg.Add(1)
		go func(t Target) {
			defer wg.Done()
			result := c.check(ctx, t)

			if err := c.store.Set(ctx, t.Name, result); err != nil {
				c.logger.Warn("failed to store result", "target", t.Name, "error", err)
			}

			if c.metrics != nil {
				c.metrics.ReportCheck(t.Name, result.Status, result.Latency.Seconds())
			}

			c.logger.Info("check complete",
				"target", t.Name,
				"status", result.Status,
				"latency", result.Latency,
			)
		}(target)
	}

	wg.Wait()
}

func (c *Checker) check(ctx context.Context, target Target) Result {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.URL, nil)
	if err != nil {
		return errResult(target, start, fmt.Sprintf("bad request: %v", err))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return errResult(target, start, fmt.Sprintf("request failed: %v", err))
	}
	defer resp.Body.Close()

	status := StatusHealthy
	if resp.StatusCode >= 400 {
		status = StatusUnhealthy
	}

	return Result{
		Target:     target.Name,
		URL:        target.URL,
		Status:     status,
		StatusCode: resp.StatusCode,
		Latency:    time.Since(start),
		CheckedAt:  time.Now(),
	}
}

func errResult(target Target, start time.Time, msg string) Result {
	return Result{
		Target:    target.Name,
		URL:       target.URL,
		Status:    StatusUnhealthy,
		Error:     msg,
		Latency:   time.Since(start),
		CheckedAt: time.Now(),
	}
}

func newCheckerFromFields(store Store, logger *slog.Logger) *Checker {
	return &Checker{
		store:  store,
		logger: logger,
	}
}
