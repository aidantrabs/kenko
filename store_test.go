package kenko

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStore_SetAndGetAll(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	r := Result{
		Target:    "api",
		URL:       "https://api.example.com",
		Status:    StatusHealthy,
		Latency:   42 * time.Millisecond,
		CheckedAt: time.Now(),
	}

	if err := s.Set(ctx, "api", r); err != nil {
		t.Fatalf("Set: %v", err)
	}

	all, err := s.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("len = %d, want 1", len(all))
	}
	if all["api"].Status != StatusHealthy {
		t.Errorf("status = %q, want %q", all["api"].Status, StatusHealthy)
	}
}

func TestMemoryStore_Overwrite(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	_ = s.Set(ctx, "api", Result{Status: StatusHealthy})
	_ = s.Set(ctx, "api", Result{Status: StatusUnhealthy})

	all, _ := s.GetAll(ctx)
	if all["api"].Status != StatusUnhealthy {
		t.Errorf("status = %q, want %q", all["api"].Status, StatusUnhealthy)
	}
}

func TestMemoryStore_GetAllEmpty(t *testing.T) {
	s := NewMemoryStore()

	all, err := s.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(all) != 0 {
		t.Fatalf("len = %d, want 0", len(all))
	}
}

func TestMemoryStore_GetAllReturnsCopy(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()
	_ = s.Set(ctx, "api", Result{Status: StatusHealthy})

	all, _ := s.GetAll(ctx)
	all["api"] = Result{Status: StatusUnhealthy}

	original, _ := s.GetAll(ctx)
	if original["api"].Status != StatusHealthy {
		t.Error("GetAll should return a copy, not a reference to internal state")
	}
}
