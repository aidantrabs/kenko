package kenko

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func testChecker() *Checker {
	store := NewMemoryStore()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return newCheckerFromFields(store, logger)
}

func TestHandleHealth_JSON(t *testing.T) {
	c := testChecker()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	HandleHealth(c)(rec, req)

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
	if resp.Status != "healthy" {
		t.Errorf("status = %q, want %q", resp.Status, "healthy")
	}
}

func TestHandleHealth_WithHealthChecker(t *testing.T) {
	store := &mockHealthStore{MemoryStore: NewMemoryStore(), pingErr: nil}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	c := newCheckerFromFields(store, logger)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	HandleHealth(c)(rec, req)

	var resp healthResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Redis != "up" {
		t.Errorf("redis = %q, want %q", resp.Redis, "up")
	}
}

func TestHandleReady_NotReady(t *testing.T) {
	c := testChecker()

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	HandleReady(c)(rec, req)

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
	c := testChecker()

	_ = c.store.Set(context.Background(), "api", Result{
		Target:    "api",
		URL:       "https://api.example.com",
		Status:    StatusHealthy,
		Latency:   42 * time.Millisecond,
		CheckedAt: time.Now(),
	})

	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	rec := httptest.NewRecorder()

	HandleStatus(c)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp statusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(resp.Targets))
	}
	if resp.Targets[0].Name != "api" {
		t.Errorf("name = %q, want %q", resp.Targets[0].Name, "api")
	}
}

type mockHealthStore struct {
	*MemoryStore
	pingErr error
}

func (m *mockHealthStore) Ping(_ context.Context) error {
	return m.pingErr
}
