package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
)

type Target struct {
	Name string
	URL  string
}

type Result struct {
	Target     string
	URL        string
	Status     Status
	StatusCode int
	Latency    time.Duration
	Error      string
	CheckedAt  time.Time
}

type Checker struct {
	client   *http.Client
	targets  []Target
	interval time.Duration
	logger   *slog.Logger

	mu      sync.RWMutex
	results map[string]Result
}

func NewChecker(targets []Target, interval, timeout time.Duration, logger *slog.Logger) *Checker {
	return &Checker{
		client:   &http.Client{Timeout: timeout},
		targets:  targets,
		interval: interval,
		logger:   logger,
		results:  make(map[string]Result),
	}
}

func (c *Checker) Run(ctx context.Context) {
	c.logger.Info("checker starting", "targets", len(c.targets), "interval", c.interval)

	c.checkAll(ctx)

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

			c.mu.Lock()
			c.results[t.Name] = result
			c.mu.Unlock()

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
		return Result{
			Target:    target.Name,
			URL:       target.URL,
			Status:    StatusUnhealthy,
			Error:     fmt.Sprintf("bad request: %v", err),
			Latency:   time.Since(start),
			CheckedAt: time.Now(),
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return Result{
			Target:    target.Name,
			URL:       target.URL,
			Status:    StatusUnhealthy,
			Error:     fmt.Sprintf("request failed: %v", err),
			Latency:   time.Since(start),
			CheckedAt: time.Now(),
		}
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

// returns a copy so callers can't corrupt internal state
func (c *Checker) Results() map[string]Result {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make(map[string]Result, len(c.results))
	for k, v := range c.results {
		out[k] = v
	}
	return out
}
