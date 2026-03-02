package monitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aidantrabs/kenko/internal/config"
)

func TestCheck_Healthy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := &Checker{
		client: ts.Client(),
	}

	target := config.Target{Name: "test", URL: ts.URL}
	result := c.check(context.Background(), target)

	if result.Status != StatusHealthy {
		t.Errorf("status = %q, want %q", result.Status, StatusHealthy)
	}
	if result.StatusCode != 200 {
		t.Errorf("status_code = %d, want 200", result.StatusCode)
	}
	if result.Error != "" {
		t.Errorf("error = %q, want empty", result.Error)
	}
}

func TestCheck_Unhealthy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := &Checker{
		client: ts.Client(),
	}

	target := config.Target{Name: "test", URL: ts.URL}
	result := c.check(context.Background(), target)

	if result.Status != StatusUnhealthy {
		t.Errorf("status = %q, want %q", result.Status, StatusUnhealthy)
	}
	if result.StatusCode != 500 {
		t.Errorf("status_code = %d, want 500", result.StatusCode)
	}
}

func TestCheck_ConnectionError(t *testing.T) {
	c := &Checker{
		client: &http.Client{},
	}

	target := config.Target{Name: "test", URL: "http://localhost:1"}
	result := c.check(context.Background(), target)

	if result.Status != StatusUnhealthy {
		t.Errorf("status = %q, want %q", result.Status, StatusUnhealthy)
	}
	if result.Error == "" {
		t.Error("expected non-empty error")
	}
}

func TestCheck_InvalidURL(t *testing.T) {
	c := &Checker{
		client: &http.Client{},
	}

	target := config.Target{Name: "test", URL: "://invalid"}
	result := c.check(context.Background(), target)

	if result.Status != StatusUnhealthy {
		t.Errorf("status = %q, want %q", result.Status, StatusUnhealthy)
	}
}

func TestReady_DefaultFalse(t *testing.T) {
	c := &Checker{}
	if c.Ready() {
		t.Error("expected Ready() = false before first check cycle")
	}
}
