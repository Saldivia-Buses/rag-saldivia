package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Resolver reads configuration from Platform DB with scope cascade.
// Resolution order: tenant:{id} > plan:{plan} > global.
// Results are cached in Redis with configurable TTL.
// Thread-safe: pgxpool and redis.Client are both safe for concurrent use.
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
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Endpoint        string  `json:"endpoint"`
	ModelID         string  `json:"model_id"`
	APIKey          string  `json:"api_key,omitempty"`
	Location        string  `json:"location"`
	CostPer1kInput  float64 `json:"cost_per_1k_input"`
	CostPer1kOutput float64 `json:"cost_per_1k_output"`
}

// cachedModelConfig is ModelConfig without the API key — safe to store in Redis.
type cachedModelConfig struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Endpoint        string  `json:"endpoint"`
	ModelID         string  `json:"model_id"`
	Location        string  `json:"location"`
	CostPer1kInput  float64 `json:"cost_per_1k_input"`
	CostPer1kOutput float64 `json:"cost_per_1k_output"`
}

// cascadeQuery resolves a config key in one round trip using a CTE.
// Checks tenant:{tenantID}, plan:{plan}, and global scopes in priority order.
const cascadeQuery = `
WITH tenant_plan AS (
    SELECT plan_id FROM tenants WHERE id = $1
)
SELECT value, scope FROM (
    SELECT value, scope, 1 AS priority
    FROM agent_config WHERE scope = 'tenant:' || $1 AND key = $2
    UNION ALL
    SELECT ac.value, ac.scope, 2 AS priority
    FROM agent_config ac, tenant_plan tp
    WHERE ac.scope = 'plan:' || tp.plan_id AND ac.key = $2
    UNION ALL
    SELECT value, scope, 3 AS priority
    FROM agent_config WHERE scope = 'global' AND key = $2
) ranked
ORDER BY priority
LIMIT 1
`

// Get resolves a config key with scope cascade in one DB query.
// tenantID can be empty (resolves global only).
func (r *Resolver) Get(ctx context.Context, tenantID, key string) (json.RawMessage, error) {
	// Try cache first
	if r.cache != nil && tenantID != "" {
		cacheKey := r.cacheKey(tenantID, key)
		val, err := r.cache.Get(ctx, cacheKey).Result()
		if err == nil {
			return json.RawMessage(val), nil
		}
	}

	var value json.RawMessage
	var scope string

	if tenantID != "" {
		// Single CTE query handles full cascade
		err := r.pool.QueryRow(ctx, cascadeQuery, tenantID, key).Scan(&value, &scope)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("config key not found: %q", key)
			}
			return nil, fmt.Errorf("query config: %w", err)
		}
	} else {
		// No tenant — global only
		err := r.pool.QueryRow(ctx,
			`SELECT value FROM agent_config WHERE scope = 'global' AND key = $1`, key,
		).Scan(&value)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("config key not found: %q", key)
			}
			return nil, fmt.Errorf("query config: %w", err)
		}
	}

	if tenantID != "" {
		r.cacheSet(ctx, tenantID, key, value)
	}
	return value, nil
}

// GetString resolves a config key and returns it as a string.
// Returns an error if the value is not a JSON string.
func (r *Resolver) GetString(ctx context.Context, tenantID, key string) (string, error) {
	raw, err := r.Get(ctx, tenantID, key)
	if err != nil {
		return "", err
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", fmt.Errorf("config %q is not a string (raw: %s): %w", key, string(raw), err)
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
		return 0, fmt.Errorf("config %q is not an integer (raw: %s): %w", key, string(raw), err)
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

	// Check model cache (cached without API key for security)
	if r.cache != nil {
		cacheKey := "sda:model:" + modelID
		val, err := r.cache.Get(ctx, cacheKey).Result()
		if err == nil {
			var mc ModelConfig
			if json.Unmarshal([]byte(val), &mc) == nil {
				// API key is never cached — fetch from DB on every call
				var apiKey string
				_ = r.pool.QueryRow(ctx,
					`SELECT COALESCE(api_key, '') FROM llm_models WHERE id = $1`,
					modelID,
				).Scan(&apiKey)
				mc.APIKey = apiKey
				return &mc, nil
			}
		}
	}

	var mc ModelConfig
	err = r.pool.QueryRow(ctx,
		`SELECT id, name, endpoint, model_id, COALESCE(api_key, ''),
			location, cost_per_1k_input, cost_per_1k_output
		 FROM llm_models WHERE id = $1 AND enabled = true`, modelID,
	).Scan(&mc.ID, &mc.Name, &mc.Endpoint, &mc.ModelID, &mc.APIKey,
		&mc.Location, &mc.CostPer1kInput, &mc.CostPer1kOutput)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("model %q not found or disabled", modelID)
		}
		return nil, fmt.Errorf("query model: %w", err)
	}

	// Cache model config without API key (security: key stays in DB only)
	if r.cache != nil {
		safe := cachedModelConfig{
			ID: mc.ID, Name: mc.Name, Endpoint: mc.Endpoint,
			ModelID: mc.ModelID, Location: mc.Location,
			CostPer1kInput: mc.CostPer1kInput, CostPer1kOutput: mc.CostPer1kOutput,
		}
		data, _ := json.Marshal(safe)
		r.cache.Set(ctx, "sda:model:"+modelID, string(data), r.cacheTTL)
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
		if errors.Is(err, pgx.ErrNoRows) {
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
	pattern := r.cacheKey(tenantID, "*")
	return r.deleteByPattern(ctx, pattern)
}

// InvalidateGlobal clears all cached config (after global config change).
func (r *Resolver) InvalidateGlobal(ctx context.Context) error {
	if r.cache == nil {
		return nil
	}
	// Clear all config cache + model cache
	if err := r.deleteByPattern(ctx, "sda:config:*"); err != nil {
		return err
	}
	return r.deleteByPattern(ctx, "sda:model:*")
}

func (r *Resolver) cacheKey(tenantID, key string) string {
	return "sda:config:" + tenantID + ":" + key
}

func (r *Resolver) cacheSet(ctx context.Context, tenantID, key string, value json.RawMessage) {
	if r.cache == nil {
		return
	}
	r.cache.Set(ctx, r.cacheKey(tenantID, key), string(value), r.cacheTTL)
}

func (r *Resolver) deleteByPattern(ctx context.Context, pattern string) error {
	var keys []string
	iter := r.cache.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return r.cache.Del(ctx, keys...).Err() // pipeline delete
	}
	return nil
}
