package redisstore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aidantrabs/kenko"
	"github.com/redis/go-redis/v9"
)

const defaultKeyPrefix = "kenko:results"

// Option configures a RedisStore.
type Option func(*RedisStore)

// WithPassword sets the Redis authentication password.
func WithPassword(password string) Option {
	return func(s *RedisStore) { s.password = password }
}

// WithKeyPrefix sets the Redis hash key used to store results (default "kenko:results").
func WithKeyPrefix(prefix string) Option {
	return func(s *RedisStore) { s.keyPrefix = prefix }
}

// RedisStore is a Store backed by a Redis hash.
type RedisStore struct {
	rdb       *redis.Client
	keyPrefix string
	password  string
}

// New creates a RedisStore connected to the given address.
func New(addr string, opts ...Option) *RedisStore {
	s := &RedisStore{keyPrefix: defaultKeyPrefix}
	for _, opt := range opts {
		opt(s)
	}
	s.rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: s.password,
	})
	return s
}

// Set stores a result in Redis keyed by target name.
func (s *RedisStore) Set(ctx context.Context, name string, result kenko.Result) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("redisstore: marshal: %w", err)
	}
	return s.rdb.HSet(ctx, s.keyPrefix, name, data).Err()
}

// GetAll retrieves all stored results from Redis.
func (s *RedisStore) GetAll(ctx context.Context) (map[string]kenko.Result, error) {
	vals, err := s.rdb.HGetAll(ctx, s.keyPrefix).Result()
	if err != nil {
		return nil, fmt.Errorf("redisstore: hgetall: %w", err)
	}

	out := make(map[string]kenko.Result, len(vals))
	for name, data := range vals {
		var r kenko.Result
		if err := json.Unmarshal([]byte(data), &r); err != nil {
			return nil, fmt.Errorf("redisstore: unmarshal %q: %w", name, err)
		}
		out[name] = r
	}
	return out, nil
}

// Ping checks connectivity to the Redis server.
func (s *RedisStore) Ping(ctx context.Context) error {
	return s.rdb.Ping(ctx).Err()
}

// Client returns the underlying Redis client.
func (s *RedisStore) Client() *redis.Client {
	return s.rdb
}
