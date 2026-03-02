package kenko

import (
	"context"
	"sync"
)

type Store interface {
	Set(ctx context.Context, name string, result Result) error
	GetAll(ctx context.Context) (map[string]Result, error)
}

// implemented by stores that can report their own health (e.g. redis ping)
type HealthChecker interface {
	Ping(ctx context.Context) error
}

type MemoryStore struct {
	mu      sync.RWMutex
	results map[string]Result
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		results: make(map[string]Result),
	}
}

func (m *MemoryStore) Set(_ context.Context, name string, result Result) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.results[name] = result
	return nil
}

func (m *MemoryStore) GetAll(_ context.Context) (map[string]Result, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make(map[string]Result, len(m.results))
	for k, v := range m.results {
		out[k] = v
	}
	return out, nil
}
