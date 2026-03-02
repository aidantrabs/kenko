package kenko

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type healthResponse struct {
	Status string `json:"status"`
	Redis  string `json:"redis,omitempty"`
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

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func HandleHealth(checker *Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := healthResponse{Status: "healthy"}

		if hc, ok := checker.store.(HealthChecker); ok {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := hc.Ping(ctx); err != nil {
				resp.Status = "degraded"
				resp.Redis = "down"
			} else {
				resp.Redis = "up"
			}
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

func HandleReady(checker *Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !checker.Ready() {
			writeJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "not_ready"})
			return
		}

		writeJSON(w, http.StatusOK, healthResponse{Status: "ready"})
	}
}

func HandleStatus(checker *Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results, err := checker.Results()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve results"})
			return
		}

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

		writeJSON(w, http.StatusOK, resp)
	}
}
