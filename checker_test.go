package kenko

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheck_Healthy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c, err := NewChecker(
		WithTarget("test", ts.URL),
		WithHTTPClient(ts.Client()),
	)
	if err != nil {
		t.Fatal(err)
	}

	result := c.check(context.Background(), Target{Name: "test", URL: ts.URL})

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

	c, err := NewChecker(
		WithTarget("test", ts.URL),
		WithHTTPClient(ts.Client()),
	)
	if err != nil {
		t.Fatal(err)
	}

	result := c.check(context.Background(), Target{Name: "test", URL: ts.URL})

	if result.Status != StatusUnhealthy {
		t.Errorf("status = %q, want %q", result.Status, StatusUnhealthy)
	}
	if result.StatusCode != 500 {
		t.Errorf("status_code = %d, want 500", result.StatusCode)
	}
}

func TestCheck_ConnectionError(t *testing.T) {
	c, err := NewChecker(
		WithTarget("test", "http://localhost:1"),
	)
	if err != nil {
		t.Fatal(err)
	}

	result := c.check(context.Background(), Target{Name: "test", URL: "http://localhost:1"})

	if result.Status != StatusUnhealthy {
		t.Errorf("status = %q, want %q", result.Status, StatusUnhealthy)
	}
	if result.Error == "" {
		t.Error("expected non-empty error")
	}
}

func TestCheck_InvalidURL(t *testing.T) {
	c, err := NewChecker(
		WithTarget("test", "http://example.com"),
	)
	if err != nil {
		t.Fatal(err)
	}

	result := c.check(context.Background(), Target{Name: "test", URL: "://invalid"})

	if result.Status != StatusUnhealthy {
		t.Errorf("status = %q, want %q", result.Status, StatusUnhealthy)
	}
}

func TestReady_DefaultFalse(t *testing.T) {
	c, _ := NewChecker(WithTarget("test", "http://example.com"))
	if c.Ready() {
		t.Error("expected Ready() = false before first check cycle")
	}
}

func TestNewChecker_NoTargets(t *testing.T) {
	_, err := NewChecker()
	if err == nil {
		t.Fatal("expected error for no targets")
	}
}

func TestNewChecker_DefaultStore(t *testing.T) {
	c, err := NewChecker(WithTarget("test", "http://example.com"))
	if err != nil {
		t.Fatal(err)
	}
	if c.Store() == nil {
		t.Error("expected non-nil default store")
	}
}
