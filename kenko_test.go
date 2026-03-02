package kenko

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew_RegistersHandlers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	k, err := New(
		WithTarget("test", ts.URL),
		WithInterval(time.Hour),
	)
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	k.RegisterHandlers(mux)

	for _, path := range []string{"/health", "/ready", "/status"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code == http.StatusNotFound {
			t.Errorf("%s returned 404", path)
		}
	}
}

func TestNew_RunAndReady(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	k, err := New(
		WithTarget("test", ts.URL),
		WithInterval(time.Hour),
	)
	if err != nil {
		t.Fatal(err)
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

	results, err := k.Checker().Results()
	if err != nil {
		t.Fatalf("Results: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("results = %d, want 1", len(results))
	}

	cancel()
	<-done
}
