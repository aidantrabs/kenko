package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aidantrabs/kenko/internal/monitor"
)

type healthResponse struct {
	Status string `json:"status"`
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
		json.NewEncoder(w).Encode(healthResponse{Status: "healthy"})
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
