package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/aidantrabs/kenko/internal/config"
	"github.com/aidantrabs/kenko/internal/monitor"
	"github.com/redis/go-redis/v9"
)

func newTestChecker() *monitor.Checker {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:1"})
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	targets := []config.Target{{Name: "test", URL: "https://example.com"}}
	return monitor.NewChecker(targets, 30*time.Second, 5*time.Second, rdb, logger)
}

func TestHandleHealth_JSON(t *testing.T) {
	checker := newTestChecker()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handleHealth(checker)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q, want application/json", ct)
	}

	var resp healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Status == "" {
		t.Error("expected non-empty status")
	}
}

func TestHandleReady_NotReady(t *testing.T) {
	checker := newTestChecker()

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	handleReady(checker)(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var resp healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Status != "not_ready" {
		t.Errorf("status = %q, want %q", resp.Status, "not_ready")
	}
}

func TestHandleStatus_JSON(t *testing.T) {
	checker := newTestChecker()

	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	rec := httptest.NewRecorder()

	handleStatus(checker)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp statusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Targets == nil {
		t.Error("expected non-nil targets slice")
	}
}
