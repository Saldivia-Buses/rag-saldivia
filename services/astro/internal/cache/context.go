package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// RedisClient is the interface for Redis operations.
// Implemented by go-redis or any compatible client.
// If nil, caching is disabled (graceful degradation).
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

// ContextCache wraps Redis for caching computed astrological contexts.
// Key pattern: astro:{tenant_id}:{contact_id}:{year}:{technique}
// TTL: natal=24h, transits=1h
// Falls through to computation if Redis is nil or unavailable.
type ContextCache struct {
	client RedisClient
}

// NewContextCache creates a context cache. Pass nil for no caching.
func NewContextCache(client RedisClient) *ContextCache {
	return &ContextCache{client: client}
}

// Available returns true if Redis is connected.
func (c *ContextCache) Available() bool {
	return c.client != nil
}

// GetJSON retrieves a cached JSON value. Returns false if not found or error.
func (c *ContextCache) GetJSON(ctx context.Context, key string, dest interface{}) bool {
	if c.client == nil {
		return false
	}
	val, err := c.client.Get(ctx, key)
	if err != nil || val == "" {
		return false
	}
	return json.Unmarshal([]byte(val), dest) == nil
}

// SetJSON stores a JSON value with TTL.
func (c *ContextCache) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	if c.client == nil {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	_ = c.client.Set(ctx, key, string(data), ttl)
}

// InvalidateContact removes all cached entries for a contact.
func (c *ContextCache) InvalidateContact(ctx context.Context, tenantID, contactID string) {
	if c.client == nil {
		return
	}
	// Delete known year range (current-2 to current+2)
	now := time.Now().Year()
	for year := now - 2; year <= now+2; year++ {
		prefix := fmt.Sprintf("astro:%s:%s:%d", tenantID, contactID, year)
		_ = c.client.Del(ctx, prefix)
	}
}

// Key builds a cache key.
func Key(tenantID, contactID string, year int, technique string) string {
	return fmt.Sprintf("astro:%s:%s:%d:%s", tenantID, contactID, year, technique)
}
