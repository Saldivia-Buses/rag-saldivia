// Package cache provides shared caching utilities for SDA services.
// Redis wrapper with graceful degradation (nil client = no-op).
// Used by astro, search, agent, and any service needing Redis JSON cache.
package cache

import (
	"context"
	"encoding/json"
	"time"
)

// RedisClient is the interface for Redis operations.
// Implemented by go-redis or any compatible client.
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

// JSONCache wraps Redis for caching JSON-serializable values.
// If client is nil, all operations are no-ops (graceful degradation).
type JSONCache struct {
	client RedisClient
}

// NewJSONCache creates a cache. Pass nil for no caching.
func NewJSONCache(client RedisClient) *JSONCache {
	return &JSONCache{client: client}
}

// Available returns true if Redis is connected.
func (c *JSONCache) Available() bool {
	return c.client != nil
}

// Get retrieves a cached JSON value. Returns false if not found or error.
func (c *JSONCache) Get(ctx context.Context, key string, dest interface{}) bool {
	if c.client == nil {
		return false
	}
	val, err := c.client.Get(ctx, key)
	if err != nil || val == "" {
		return false
	}
	return json.Unmarshal([]byte(val), dest) == nil
}

// Set stores a JSON value with TTL.
func (c *JSONCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	if c.client == nil {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	_ = c.client.Set(ctx, key, string(data), ttl)
}

// Del deletes one or more keys.
func (c *JSONCache) Del(ctx context.Context, keys ...string) {
	if c.client == nil {
		return
	}
	_ = c.client.Del(ctx, keys...)
}
