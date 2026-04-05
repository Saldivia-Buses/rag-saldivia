package config

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Resolver reads configuration from Platform DB with scope cascade.
// Resolution order: tenant:{id} > plan:{plan} > global.
// Results are cached in Redis with configurable TTL.
type Resolver struct {
	pool     *pgxpool.Pool
	cache    *redis.Client
	cacheTTL time.Duration
}

// NewResolver creates a config resolver.
// cache may be nil (no caching). Default TTL is 5 minutes.
func NewResolver(pool *pgxpool.Pool, cache *redis.Client) *Resolver {
	return &Resolver{
		pool:     pool,
		cache:    cache,
		cacheTTL: 5 * time.Minute,
	}
}

// ModelConfig is a resolved LLM model configuration from llm_models.
type ModelConfig struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Endpoint          string  `json:"endpoint"`
	ModelID           string  `json:"model_id"`
	APIKey            string  `json:"api_key,omitempty"`
	Location          string  `json:"location"`
	CostPer1kInput    float64 `json:"cost_per_1k_input"`
	CostPer1kOutput   float64 `json:"cost_per_1k_output"`
}

// Get resolves a config key with scope cascade.
// Checks tenant:{tenantID} first, then plan:{plan}, then global.
// tenantID can be empty (resolves global only).
func (r *Resolver) Get(ctx context.Context, tenantID, key string) (json.RawMessage, error) {
	// Try cache first
	if r.cache != nil && tenantID != "" {
		cacheKey := "sda:config:" + tenantID + ":" + key
		val, err := r.cache.Get(ctx, cacheKey).Result()
		if err == nil {
			return json.RawMessage(val), nil
		}
	}

	// Cascade: tenant > plan > global
	var value json.RawMessage

	if tenantID != "" {
		// Try tenant-specific
		err := r.pool.QueryRow(ctx,
			`SELECT value FROM agent_config WHERE scope = $1 AND key = $2`,
			"tenant:"+tenantID, key,
		).Scan(&value)
		if err == nil {
			r.cacheSet(ctx, tenantID, key, value)
			return value, nil
		}

		// Try plan-specific (need to look up tenant's plan)
		var planID string
		err = r.pool.QueryRow(ctx,
			`SELECT plan_id FROM tenants WHERE id = $1`, tenantID,
		).Scan(&planID)
		if err == nil && planID != "" {
			err = r.pool.QueryRow(ctx,
				`SELECT value FROM agent_config WHERE scope = $1 AND key = $2`,
				"plan:"+planID, key,
			).Scan(&value)
			if err == nil {
				r.cacheSet(ctx, tenantID, key, value)
				return value, nil
			}
		}
	}

	// Try global
	err := r.pool.QueryRow(ctx,
		`SELECT value FROM agent_config WHERE scope = 'global' AND key = $1`, key,
	).Scan(&value)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("config key not found: %q", key)
		}
		return nil, fmt.Errorf("query config: %w", err)
	}

	if tenantID != "" {
		r.cacheSet(ctx, tenantID, key, value)
	}
	return value, nil
}

// GetString resolves a config key and returns it as a string.
func (r *Resolver) GetString(ctx context.Context, tenantID, key string) (string, error) {
	raw, err := r.Get(ctx, tenantID, key)
	if err != nil {
		return "", err
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return string(raw), nil // return raw if not a quoted string
	}
	return s, nil
}

// GetInt resolves a config key and returns it as an int.
func (r *Resolver) GetInt(ctx context.Context, tenantID, key string) (int, error) {
	raw, err := r.Get(ctx, tenantID, key)
	if err != nil {
		return 0, err
	}
	var n int
	if err := json.Unmarshal(raw, &n); err != nil {
		return 0, fmt.Errorf("config %q is not an integer: %w", key, err)
	}
	return n, nil
}

// ResolveSlot resolves a pipeline slot to its model configuration.
// Reads slot.{name} from agent_config → gets model ID → looks up llm_models.
func (r *Resolver) ResolveSlot(ctx context.Context, tenantID, slot string) (*ModelConfig, error) {
	modelID, err := r.GetString(ctx, tenantID, slot)
	if err != nil {
		return nil, fmt.Errorf("resolve slot %q: %w", slot, err)
	}
	if modelID == "" || modelID == "null" {
		return nil, fmt.Errorf("slot %q is not configured", slot)
	}

	var mc ModelConfig
	err = r.pool.QueryRow(ctx,
		`SELECT id, name, endpoint, model_id, COALESCE(api_key, ''),
			location, cost_per_1k_input, cost_per_1k_output
		 FROM llm_models WHERE id = $1 AND enabled = true`, modelID,
	).Scan(&mc.ID, &mc.Name, &mc.Endpoint, &mc.ModelID, &mc.APIKey,
		&mc.Location, &mc.CostPer1kInput, &mc.CostPer1kOutput)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("model %q not found or disabled", modelID)
		}
		return nil, fmt.Errorf("query model: %w", err)
	}

	return &mc, nil
}

// GetActivePrompt returns the active prompt content for a key.
func (r *Resolver) GetActivePrompt(ctx context.Context, promptKey string) (string, error) {
	var content string
	err := r.pool.QueryRow(ctx,
		`SELECT content FROM prompt_versions
		 WHERE prompt_key = $1 AND is_active = true
		 ORDER BY version DESC LIMIT 1`, promptKey,
	).Scan(&content)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("no active prompt for key %q", promptKey)
		}
		return "", fmt.Errorf("query prompt: %w", err)
	}
	return content, nil
}

// InvalidateCache clears cached config for a tenant.
func (r *Resolver) InvalidateCache(ctx context.Context, tenantID string) error {
	if r.cache == nil {
		return nil
	}
	pattern := "sda:config:" + tenantID + ":*"
	iter := r.cache.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		r.cache.Del(ctx, iter.Val())
	}
	return iter.Err()
}

func (r *Resolver) cacheSet(ctx context.Context, tenantID, key string, value json.RawMessage) {
	if r.cache == nil {
		return
	}
	cacheKey := "sda:config:" + tenantID + ":" + key
	r.cache.Set(ctx, cacheKey, string(value), r.cacheTTL)
}
