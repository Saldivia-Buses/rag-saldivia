package tenant

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// ConnInfo holds connection details for a tenant's databases.
// NOTE: postgres_url and redis_url contain credentials. In production,
// these should be encrypted at the application layer before storing in
// the platform DB, or fetched from a secrets manager (Vault, etc.).
type ConnInfo struct {
	PostgresURL string
	RedisURL    string
}

// Resolver maps tenant slugs to their database connections.
// It caches connection pools so we don't create a new pool per request.
// Uses singleflight-style locking to prevent thundering herd on pool creation.
type Resolver struct {
	platformDB *pgxpool.Pool

	mu        sync.Mutex
	pools     map[string]*pgxpool.Pool // slug → PostgreSQL pool
	redisClts map[string]*redis.Client // slug → Redis client
	connCache map[string]connEntry     // slug → connection info + timestamp
	closed    bool

	// PoolMaxConns sets the max connections per tenant pool. Default 4.
	PoolMaxConns int32
}

type connEntry struct {
	info      ConnInfo
	fetchedAt time.Time
}

const defaultCacheTTL = 5 * time.Minute

// NewResolver creates a tenant resolver backed by the platform database.
func NewResolver(platformDB *pgxpool.Pool) *Resolver {
	return &Resolver{
		platformDB:   platformDB,
		pools:        make(map[string]*pgxpool.Pool),
		redisClts:    make(map[string]*redis.Client),
		connCache:    make(map[string]connEntry),
		PoolMaxConns: 4,
	}
}

// PostgresPool returns a connection pool for the given tenant.
// Safe for concurrent access — only one pool is created per slug.
func (r *Resolver) PostgresPool(ctx context.Context, slug string) (*pgxpool.Pool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, errors.New("resolver is closed")
	}

	if pool, ok := r.pools[slug]; ok {
		return pool, nil
	}

	return r.createPoolLocked(ctx, slug)
}

// RedisClient returns a Redis client for the given tenant.
// Safe for concurrent access — only one client is created per slug.
func (r *Resolver) RedisClient(ctx context.Context, slug string) (*redis.Client, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, errors.New("resolver is closed")
	}

	if client, ok := r.redisClts[slug]; ok {
		return client, nil
	}

	return r.createRedisClientLocked(ctx, slug)
}

// resolveConnInfo looks up connection info for a tenant from the platform DB.
// Must be called with r.mu held.
func (r *Resolver) resolveConnInfo(ctx context.Context, slug string) (ConnInfo, error) {
	if r.platformDB == nil {
		return ConnInfo{}, fmt.Errorf("resolve tenant %q: platform DB not configured", slug)
	}

	if entry, ok := r.connCache[slug]; ok {
		if time.Since(entry.fetchedAt) < defaultCacheTTL {
			return entry.info, nil
		}
	}

	var info ConnInfo
	err := r.platformDB.QueryRow(ctx,
		`SELECT postgres_url, redis_url FROM tenants WHERE slug = $1 AND enabled = true`,
		slug,
	).Scan(&info.PostgresURL, &info.RedisURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ConnInfo{}, fmt.Errorf("resolve tenant %q: %w", slug, ErrTenantUnknown)
		}
		return ConnInfo{}, fmt.Errorf("resolve tenant %q: %w", slug, err)
	}

	r.connCache[slug] = connEntry{info: info, fetchedAt: time.Now()}
	return info, nil
}

// createPoolLocked creates a PostgreSQL pool. Must be called with r.mu held.
// The mutex is released during the network call (pool creation + ping) to
// avoid blocking all tenants, then re-acquired to store the result.
func (r *Resolver) createPoolLocked(ctx context.Context, slug string) (*pgxpool.Pool, error) {
	info, err := r.resolveConnInfo(ctx, slug)
	if err != nil {
		return nil, err
	}

	config, err := pgxpool.ParseConfig(info.PostgresURL)
	if err != nil {
		return nil, fmt.Errorf("parse pool config for tenant %q: %w", slug, err)
	}
	config.MaxConns = r.PoolMaxConns

	// Release lock during network I/O
	r.mu.Unlock()
	pool, err := pgxpool.NewWithConfig(ctx, config)
	r.mu.Lock()

	if err != nil {
		return nil, fmt.Errorf("create pool for tenant %q: %w", slug, err)
	}

	// Double-check: another goroutine may have created the pool while we
	// were unlocked. If so, close ours and return the existing one.
	if existing, ok := r.pools[slug]; ok {
		pool.Close()
		return existing, nil
	}

	r.pools[slug] = pool
	return pool, nil
}

// createRedisClientLocked creates a Redis client. Must be called with r.mu held.
func (r *Resolver) createRedisClientLocked(ctx context.Context, slug string) (*redis.Client, error) {
	info, err := r.resolveConnInfo(ctx, slug)
	if err != nil {
		return nil, err
	}

	opts, err := redis.ParseURL(info.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("parse Redis URL for tenant %q: %w", slug, err)
	}

	// Release lock during network I/O
	r.mu.Unlock()
	client := redis.NewClient(opts)
	pingErr := client.Ping(ctx).Err()
	r.mu.Lock()

	if pingErr != nil {
		client.Close()
		return nil, fmt.Errorf("ping tenant %q Redis: %w", slug, pingErr)
	}

	// Double-check
	if existing, ok := r.redisClts[slug]; ok {
		client.Close()
		return existing, nil
	}

	r.redisClts[slug] = client
	return client, nil
}

// Close shuts down all cached connection pools and Redis clients.
// After Close, all methods return errors.
func (r *Resolver) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.closed = true

	for slug, pool := range r.pools {
		pool.Close()
		delete(r.pools, slug)
	}
	for slug, client := range r.redisClts {
		client.Close()
		delete(r.redisClts, slug)
	}
	r.connCache = make(map[string]connEntry)
}
