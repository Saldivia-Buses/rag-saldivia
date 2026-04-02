package tenant

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// ConnInfo holds connection details for a tenant's databases.
type ConnInfo struct {
	PostgresURL string
	RedisURL    string
}

// Resolver maps tenant slugs to their database connections.
// It caches connection pools so we don't create a new pool per request.
type Resolver struct {
	platformDB *pgxpool.Pool
	mu         sync.RWMutex
	pools      map[string]*pgxpool.Pool   // slug → PostgreSQL pool
	redisClts  map[string]*redis.Client   // slug → Redis client
	connCache  map[string]ConnInfo        // slug → connection info (from platform DB)
	cacheTTL   time.Duration
	cacheTime  map[string]time.Time       // slug → when cached
}

// NewResolver creates a tenant resolver backed by the platform database.
// The platform DB is used to look up connection info for each tenant.
func NewResolver(platformDB *pgxpool.Pool) *Resolver {
	return &Resolver{
		platformDB: platformDB,
		pools:      make(map[string]*pgxpool.Pool),
		redisClts:  make(map[string]*redis.Client),
		connCache:  make(map[string]ConnInfo),
		cacheTTL:   5 * time.Minute,
		cacheTime:  make(map[string]time.Time),
	}
}

// PostgresPool returns a connection pool for the given tenant.
// The pool is created on first access and cached for subsequent requests.
func (r *Resolver) PostgresPool(ctx context.Context, slug string) (*pgxpool.Pool, error) {
	r.mu.RLock()
	if pool, ok := r.pools[slug]; ok {
		r.mu.RUnlock()
		return pool, nil
	}
	r.mu.RUnlock()

	return r.createPool(ctx, slug)
}

// RedisClient returns a Redis client for the given tenant.
func (r *Resolver) RedisClient(ctx context.Context, slug string) (*redis.Client, error) {
	r.mu.RLock()
	if client, ok := r.redisClts[slug]; ok {
		r.mu.RUnlock()
		return client, nil
	}
	r.mu.RUnlock()

	return r.createRedisClient(ctx, slug)
}

// resolveConnInfo looks up connection info for a tenant from the platform DB.
// Results are cached for cacheTTL to avoid hitting the platform DB on every request.
func (r *Resolver) resolveConnInfo(ctx context.Context, slug string) (ConnInfo, error) {
	r.mu.RLock()
	if info, ok := r.connCache[slug]; ok {
		if time.Since(r.cacheTime[slug]) < r.cacheTTL {
			r.mu.RUnlock()
			return info, nil
		}
	}
	r.mu.RUnlock()

	var info ConnInfo
	err := r.platformDB.QueryRow(ctx,
		`SELECT postgres_url, redis_url FROM tenants WHERE slug = $1 AND enabled = true`,
		slug,
	).Scan(&info.PostgresURL, &info.RedisURL)
	if err != nil {
		return ConnInfo{}, fmt.Errorf("resolve tenant %q: %w", slug, ErrTenantUnknown)
	}

	r.mu.Lock()
	r.connCache[slug] = info
	r.cacheTime[slug] = time.Now()
	r.mu.Unlock()

	return info, nil
}

func (r *Resolver) createPool(ctx context.Context, slug string) (*pgxpool.Pool, error) {
	info, err := r.resolveConnInfo(ctx, slug)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.New(ctx, info.PostgresURL)
	if err != nil {
		return nil, fmt.Errorf("create pool for tenant %q: %w", slug, err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping tenant %q DB: %w", slug, err)
	}

	r.mu.Lock()
	r.pools[slug] = pool
	r.mu.Unlock()

	return pool, nil
}

func (r *Resolver) createRedisClient(ctx context.Context, slug string) (*redis.Client, error) {
	info, err := r.resolveConnInfo(ctx, slug)
	if err != nil {
		return nil, err
	}

	opts, err := redis.ParseURL(info.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("parse Redis URL for tenant %q: %w", slug, err)
	}

	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping tenant %q Redis: %w", slug, err)
	}

	r.mu.Lock()
	r.redisClts[slug] = client
	r.mu.Unlock()

	return client, nil
}

// Close shuts down all cached connection pools and Redis clients.
func (r *Resolver) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, pool := range r.pools {
		pool.Close()
	}
	for _, client := range r.redisClts {
		client.Close()
	}
}
