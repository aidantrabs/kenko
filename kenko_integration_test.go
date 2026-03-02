package kenko

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIntegration_FullFlow(t *testing.T) {
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer healthy.Close()

	unhealthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer unhealthy.Close()

	k, err := New(
		WithTarget("up", healthy.URL),
		WithTarget("down", unhealthy.URL),
		WithInterval(time.Hour), // long interval so only the initial check runs
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	mux := http.NewServeMux()
	k.RegisterHandlers(mux)

	if k.Checker().Ready() {
		t.Fatal("expected Ready()=false before Run")
	}

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("/ready before Run: got %d, want 503", rec.Code)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		k.Run(ctx)
		close(done)
	}()

	deadline := time.After(5 * time.Second)
	for !k.Checker().Ready() {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for ready")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("/ready: got %d, want 200", rec.Code)
	}
	var readyResp healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&readyResp); err != nil {
		t.Fatalf("/ready decode: %v", err)
	}
	if readyResp.Status != "ready" {
		t.Errorf("/ready status = %q, want %q", readyResp.Status, "ready")
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("/health: got %d, want 200", rec.Code)
	}
	var healthResp healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&healthResp); err != nil {
		t.Fatalf("/health decode: %v", err)
	}
	if healthResp.Status != "healthy" {
		t.Errorf("/health status = %q, want %q", healthResp.Status, "healthy")
	}
	if healthResp.Redis != "" {
		t.Errorf("/health redis = %q, want empty (MemoryStore)", healthResp.Redis)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/status", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("/status: got %d, want 200", rec.Code)
	}
	var statusResp statusResponse
	if err := json.NewDecoder(rec.Body).Decode(&statusResp); err != nil {
		t.Fatalf("/status decode: %v", err)
	}
	if len(statusResp.Targets) != 2 {
		t.Fatalf("/status targets = %d, want 2", len(statusResp.Targets))
	}

	byName := make(map[string]targetResult)
	for _, tr := range statusResp.Targets {
		byName[tr.Name] = tr
	}

	up, ok := byName["up"]
	if !ok {
		t.Fatal("/status missing target 'up'")
	}
	if up.Status != "healthy" || up.StatusCode != 200 || up.URL != healthy.URL {
		t.Errorf("up target: status=%q code=%d url=%q", up.Status, up.StatusCode, up.URL)
	}

	down, ok := byName["down"]
	if !ok {
		t.Fatal("/status missing target 'down'")
	}
	if down.Status != "unhealthy" || down.StatusCode != 500 || down.URL != unhealthy.URL {
		t.Errorf("down target: status=%q code=%d url=%q", down.Status, down.StatusCode, down.URL)
	}

	results, err := k.Checker().Results()
	if err != nil {
		t.Fatalf("Results: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Results len = %d, want 2", len(results))
	}
	if results["up"].Status != StatusHealthy {
		t.Errorf("Results[up].Status = %q, want %q", results["up"].Status, StatusHealthy)
	}
	if results["down"].Status != StatusUnhealthy {
		t.Errorf("Results[down].Status = %q, want %q", results["down"].Status, StatusUnhealthy)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for shutdown")
	}
}
