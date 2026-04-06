package tenant

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	sdacrypto "github.com/Camionerou/rag-saldivia/pkg/crypto"
)

// ConnInfo holds connection details for a tenant's databases.
// NOTE: postgres_url and redis_url contain credentials. In production,
// these should be encrypted at the application layer before storing in
// the platform DB, or fetched from a secrets manager (Vault, etc.).
type ConnInfo struct {
	TenantID    string // UUID from platform DB
	PostgresURL string
	RedisURL    string
}

// Resolver maps tenant slugs to their database connections.
// It caches connection pools so we don't create a new pool per request.
// Uses singleflight-style locking to prevent thundering herd on pool creation.
type Resolver struct {
	platformDB    *pgxpool.Pool
	encryptionKey []byte // 32-byte AES-256 key; nil = no encryption (backwards compat)

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
// encryptionKey is optional (nil = credentials stored in plaintext).
func NewResolver(platformDB *pgxpool.Pool, encryptionKey []byte) *Resolver {
	return &Resolver{
		platformDB:    platformDB,
		encryptionKey: encryptionKey,
		pools:         make(map[string]*pgxpool.Pool),
		redisClts:     make(map[string]*redis.Client),
		connCache:     make(map[string]connEntry),
		PoolMaxConns:  4, // override via Resolver.PoolMaxConns before first use
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
	var pgURL, redisURL string
	var pgURLEnc, redisURLEnc *string
	err := r.platformDB.QueryRow(ctx,
		`SELECT id, postgres_url, redis_url,
		        postgres_url_enc, redis_url_enc
		 FROM tenants WHERE slug = $1 AND enabled = true`,
		slug,
	).Scan(&info.TenantID, &pgURL, &redisURL, &pgURLEnc, &redisURLEnc)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ConnInfo{}, fmt.Errorf("resolve tenant %q: %w", slug, ErrTenantUnknown)
		}
		// Fallback: _enc columns may not exist yet (pre-migration)
		err2 := r.platformDB.QueryRow(ctx,
			`SELECT id, postgres_url, redis_url FROM tenants WHERE slug = $1 AND enabled = true`,
			slug,
		).Scan(&info.TenantID, &info.PostgresURL, &info.RedisURL)
		if err2 != nil {
			if errors.Is(err2, pgx.ErrNoRows) {
				return ConnInfo{}, fmt.Errorf("resolve tenant %q: %w", slug, ErrTenantUnknown)
			}
			return ConnInfo{}, fmt.Errorf("resolve tenant %q: %w", slug, err2)
		}
		r.connCache[slug] = connEntry{info: info, fetchedAt: time.Now()}
		return info, nil
	}

	// Prefer encrypted URLs if available and we have a key
	if r.encryptionKey != nil && pgURLEnc != nil && *pgURLEnc != "" {
		decrypted, err := sdacrypto.Decrypt(r.encryptionKey, *pgURLEnc)
		if err != nil {
			return ConnInfo{}, fmt.Errorf("decrypt postgres_url for tenant %q: %w", slug, err)
		}
		info.PostgresURL = decrypted
	} else {
		info.PostgresURL = pgURL
	}

	if r.encryptionKey != nil && redisURLEnc != nil && *redisURLEnc != "" {
		decrypted, err := sdacrypto.Decrypt(r.encryptionKey, *redisURLEnc)
		if err != nil {
			return ConnInfo{}, fmt.Errorf("decrypt redis_url for tenant %q: %w", slug, err)
		}
		info.RedisURL = decrypted
	} else {
		info.RedisURL = redisURL
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

	pgURL := ensureSSLMode(info.PostgresURL)
	config, err := pgxpool.ParseConfig(pgURL)
	if err != nil {
		return nil, fmt.Errorf("parse pool config for tenant %q: %w", slug, err)
	}
	config.MaxConns = r.PoolMaxConns

	// Release lock during network I/O. Recover wrapper protects mutex
	// state if pool creation panics unexpectedly.
	r.mu.Unlock()
	pool, err := func() (p *pgxpool.Pool, createErr error) {
		defer func() {
			if rv := recover(); rv != nil {
				createErr = fmt.Errorf("pool creation panicked: %v", rv)
			}
		}()
		return pgxpool.NewWithConfig(ctx, config)
	}()
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

// TenantID returns the UUID for the given tenant slug (cached).
func (r *Resolver) TenantID(ctx context.Context, slug string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	info, err := r.resolveConnInfo(ctx, slug)
	if err != nil {
		return "", err
	}
	return info.TenantID, nil
}

// EnabledModule represents a module enabled for a tenant.
type EnabledModule struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

// ListEnabledModules returns the modules enabled for a tenant by querying
// the Platform DB. Falls back to core modules if the query fails.
func (r *Resolver) ListEnabledModules(ctx context.Context, tenantID string) ([]EnabledModule, error) {
	if r.platformDB == nil {
		return coreModules(), nil
	}

	rows, err := r.platformDB.Query(ctx,
		`SELECT m.id, m.name, m.category
		 FROM tenant_modules tm
		 JOIN modules m ON m.id = tm.module_id
		 WHERE tm.tenant_id = $1 AND tm.enabled = true
		 UNION
		 SELECT id, name, category FROM modules WHERE category = 'core'
		 ORDER BY category, name`,
		tenantID,
	)
	if err != nil {
		slog.Warn("failed to query enabled modules, using core defaults", "error", err)
		return coreModules(), nil
	}
	defer rows.Close()

	var modules []EnabledModule
	for rows.Next() {
		var m EnabledModule
		if err := rows.Scan(&m.ID, &m.Name, &m.Category); err != nil {
			continue
		}
		modules = append(modules, m)
	}
	if len(modules) == 0 {
		return coreModules(), nil
	}
	return modules, nil
}

func coreModules() []EnabledModule {
	return []EnabledModule{
		{ID: "chat", Name: "Chat + RAG", Category: "core"},
		{ID: "auth", Name: "Auth + RBAC", Category: "core"},
		{ID: "notifications", Name: "Notificaciones", Category: "core"},
	}
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

// StartHealthCheck launches a background goroutine that pings all cached
// pools every interval. Unhealthy pools are closed and removed from the
// cache so the next request creates a fresh connection.
func (r *Resolver) StartHealthCheck(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.mu.Lock()
				for slug, pool := range r.pools {
					if err := pool.Ping(ctx); err != nil {
						slog.Warn("tenant pool unhealthy, removing from cache",
							"tenant", slug, "error", err)
						pool.Close()
						delete(r.pools, slug)
						delete(r.connCache, slug)
					}
				}
				r.mu.Unlock()
			}
		}
	}()
}

// ensureSSLMode appends sslmode=require to a PostgreSQL URL if no sslmode
// is already specified. This enforces encrypted connections in production
// without breaking dev URLs that explicitly set sslmode=disable.
func ensureSSLMode(pgURL string) string {
	u, err := url.Parse(pgURL)
	if err != nil {
		slog.Warn("failed to parse PG URL for SSL enforcement", "error", err)
		return pgURL
	}
	q := u.Query()
	if q.Get("sslmode") == "" {
		q.Set("sslmode", "require")
		u.RawQuery = q.Encode()
		return u.String()
	}
	return pgURL
}
