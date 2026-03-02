package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aidantrabs/kenko/internal/monitor"
)

type healthResponse struct {
	Status string `json:"status"`
	Redis  string `json:"redis"`
}

type statusResponse struct {
	Targets []targetResult `json:"targets"`
}

type targetResult struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	Status     string `json:"status"`
	StatusCode int    `json:"status_code,omitempty"`
	LatencyMS  int64  `json:"latency_ms"`
	Error      string `json:"error,omitempty"`
	CheckedAt  string `json:"checked_at"`
}

func handleHealth(checker *monitor.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		redisStatus := "up"
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := checker.RedisClient().Ping(ctx).Err(); err != nil {
			redisStatus = "down"
		}

		status := "healthy"
		if redisStatus == "down" {
			status = "degraded"
		}

		json.NewEncoder(w).Encode(healthResponse{
			Status: status,
			Redis:  redisStatus,
		})
	}
}

func handleReady(checker *monitor.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !checker.Ready() {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(healthResponse{Status: "not_ready", Redis: "unknown"})
			return
		}

		json.NewEncoder(w).Encode(healthResponse{Status: "ready", Redis: "up"})
	}
}

func handleStatus(checker *monitor.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		results := checker.Results()
		resp := statusResponse{
			Targets: make([]targetResult, 0, len(results)),
		}

		for _, r := range results {
			resp.Targets = append(resp.Targets, targetResult{
				Name:       r.Target,
				URL:        r.URL,
				Status:     string(r.Status),
				StatusCode: r.StatusCode,
				LatencyMS:  r.Latency.Milliseconds(),
				Error:      r.Error,
				CheckedAt:  r.CheckedAt.Format(time.RFC3339),
			})
		}

		json.NewEncoder(w).Encode(resp)
	}
}
