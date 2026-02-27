package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/aidantrabs/kenko/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"

	redisKey = "kenko:results"
)

type Result struct {
	Target     string        `json:"target"`
	URL        string        `json:"url"`
	Status     Status        `json:"status"`
	StatusCode int           `json:"status_code"`
	Latency    time.Duration `json:"latency"`
	Error      string        `json:"error,omitempty"`
	CheckedAt  time.Time     `json:"checked_at"`
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
	rdb      *redis.Client
	targets  []config.Target
	interval time.Duration
	logger   *slog.Logger

	mu      sync.RWMutex
	results map[string]Result
}

func NewChecker(targets []config.Target, interval, timeout time.Duration, rdb *redis.Client, logger *slog.Logger) *Checker {
	return &Checker{
		client:   &http.Client{Timeout: timeout},
		rdb:      rdb,
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

			c.storeResult(ctx, t.Name, result)

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

func (c *Checker) storeResult(ctx context.Context, name string, result Result) {
	data, err := json.Marshal(result)
	if err != nil {
		c.logger.Error("failed to marshal result", "target", name, "error", err)
		return
	}

	if err := c.rdb.HSet(ctx, redisKey, name, data).Err(); err != nil {
		c.logger.Warn("failed to write to redis", "target", name, "error", err)
	}
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

// reads from redis first, falls back to in-memory
func (c *Checker) Results() map[string]Result {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	vals, err := c.rdb.HGetAll(ctx, redisKey).Result()
	if err == nil && len(vals) > 0 {
		out := make(map[string]Result, len(vals))
		for name, data := range vals {
			var r Result
			if err := json.Unmarshal([]byte(data), &r); err == nil {
				out[name] = r
			}
		}
		return out
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make(map[string]Result, len(c.results))
	for k, v := range c.results {
		out[k] = v
	}
	return out
}
