package kenko

import (
	"context"
	"sync"
)

// Store persists and retrieves health check results.
type Store interface {
	Set(ctx context.Context, name string, result Result) error
	GetAll(ctx context.Context) (map[string]Result, error)
}

// HealthChecker is implemented by stores that can report their own health (e.g. Redis ping).
type HealthChecker interface {
	Ping(ctx context.Context) error
}

// MemoryStore is an in-memory Store implementation safe for concurrent use.
type MemoryStore struct {
	mu      sync.RWMutex
	results map[string]Result
}

// NewMemoryStore returns an initialized MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		results: make(map[string]Result),
	}
}

// Set stores a result keyed by target name.
func (m *MemoryStore) Set(_ context.Context, name string, result Result) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.results[name] = result
	return nil
}

// GetAll returns a copy of all stored results.
func (m *MemoryStore) GetAll(_ context.Context) (map[string]Result, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make(map[string]Result, len(m.results))
	for k, v := range m.results {
		out[k] = v
	}
	return out, nil
}
