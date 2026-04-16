// Package featureflags provides a client to evaluate feature flags
// from the platform service. It caches results with a configurable TTL.
package featureflags

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const maxCacheEntries = 10000

// Client queries the platform service for evaluated feature flags.
type Client struct {
	platformURL string
	httpClient  *http.Client
	cacheTTL    time.Duration

	mu    sync.RWMutex
	cache map[string]cacheEntry
}

type cacheEntry struct {
	flags   map[string]bool
	expires time.Time
}

type evaluateResponse struct {
	Flags map[string]bool `json:"flags"`
}

// New creates a feature flags client.
// platformURL should be the base URL of the platform service (e.g. "http://localhost:8006").
func New(platformURL string) *Client {
	return &Client{
		platformURL: platformURL,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		cacheTTL:    30 * time.Second,
		cache:       make(map[string]cacheEntry),
	}
}

// IsEnabled returns whether a flag is enabled for the caller identified by the JWT.
// The jwt parameter is forwarded as Authorization header to the platform service.
func (c *Client) IsEnabled(ctx context.Context, flag string, jwt string) bool {
	flags, err := c.evaluate(ctx, jwt)
	if err != nil {
		return false
	}
	return flags[flag]
}

// evaluate fetches and caches all flags for the given JWT.
// Cache key is a SHA-256 hash of the JWT (never stores raw tokens in memory).
func (c *Client) evaluate(ctx context.Context, jwt string) (map[string]bool, error) {
	h := sha256.Sum256([]byte(jwt))
	cacheKey := hex.EncodeToString(h[:16])

	c.mu.RLock()
	if entry, ok := c.cache[cacheKey]; ok && time.Now().Before(entry.expires) {
		c.mu.RUnlock()
		return entry.flags, nil
	}
	c.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.platformURL+"/v1/flags/evaluate", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwt)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("evaluate flags: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("evaluate flags: status %d", resp.StatusCode)
	}

	var result evaluateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode flags: %w", err)
	}

	c.mu.Lock()
	if len(c.cache) >= maxCacheEntries {
		c.cache = make(map[string]cacheEntry)
	}
	c.cache[cacheKey] = cacheEntry{flags: result.Flags, expires: time.Now().Add(c.cacheTTL)}
	c.mu.Unlock()

	return result.Flags, nil
}
