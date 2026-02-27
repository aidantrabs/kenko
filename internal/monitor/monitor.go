package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/aidantrabs/kenko/internal/config"
	"github.com/prometheus/client_golang/prometheus"
)

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
)

type Result struct {
	Target     string
	URL        string
	Status     Status
	StatusCode int
	Latency    time.Duration
	Error      string
	CheckedAt  time.Time
}

var (
	checkDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kenko_check_duration_seconds",
		Help:    "duration of health checks",
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"target"})

	checkTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kenko_check_total",
		Help: "total number of health checks",
	}, []string{"target", "status"})

	targetUp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kenko_target_up",
		Help: "whether a target is healthy (1) or not (0)",
	}, []string{"target"})
)

func init() {
	prometheus.MustRegister(checkDuration, checkTotal, targetUp)
}

type Checker struct {
	client   *http.Client
	targets  []config.Target
	interval time.Duration
	logger   *slog.Logger

	mu      sync.RWMutex
	results map[string]Result
}

func NewChecker(targets []config.Target, interval, timeout time.Duration, logger *slog.Logger) *Checker {
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
		go func(t config.Target) {
			defer wg.Done()
			result := c.check(ctx, t)

			c.mu.Lock()
			c.results[t.Name] = result
			c.mu.Unlock()

			checkDuration.WithLabelValues(t.Name).Observe(result.Latency.Seconds())
			checkTotal.WithLabelValues(t.Name, string(result.Status)).Inc()
			if result.Status == StatusHealthy {
				targetUp.WithLabelValues(t.Name).Set(1)
			} else {
				targetUp.WithLabelValues(t.Name).Set(0)
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

func (c *Checker) check(ctx context.Context, target config.Target) Result {
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
