package outbox

import (
	"context"
	"log/slog"
	"sync"

	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/services/app/internal/spine"
)

// DrainerRegistry manages one DrainerWorker per active tenant. It bootstraps
// from the list of known tenants at startup and can hotload new tenants via
// AddTenant / RemoveTenant (called by a platform.lifecycle consumer).
type DrainerRegistry struct {
	mu      sync.Mutex
	pool    spine.TenantPool
	nc      *nats.Conn
	workers map[string]context.CancelFunc
	opts    []DrainerOpt
}

// NewRegistry creates a registry. Call Start to bootstrap from known tenants,
// then AddTenant / RemoveTenant as platform.lifecycle events arrive.
func NewRegistry(pool spine.TenantPool, nc *nats.Conn, opts ...DrainerOpt) *DrainerRegistry {
	return &DrainerRegistry{
		pool:    pool,
		nc:      nc,
		workers: make(map[string]context.CancelFunc),
		opts:    opts,
	}
}

// Start bootstraps drainers for all known tenants. Blocks until ctx is done.
// Call this as a goroutine from the service's main.
func (r *DrainerRegistry) Start(ctx context.Context, tenantSlugs []string) {
	for _, slug := range tenantSlugs {
		r.AddTenant(ctx, slug)
	}
	<-ctx.Done()
	r.stopAll()
}

// AddTenant starts a drainer for a new tenant. Safe to call if the tenant
// already has a drainer (no-op).
func (r *DrainerRegistry) AddTenant(ctx context.Context, slug string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.workers[slug]; exists {
		return
	}

	tenantPool, err := r.pool.PostgresPool(ctx, slug)
	if err != nil {
		slog.Error("outbox: cannot start drainer — pool resolve failed",
			"tenant", slug, "error", err)
		return
	}

	workerCtx, cancel := context.WithCancel(ctx)
	r.workers[slug] = cancel

	drainer := NewDrainer(tenantPool, r.nc, slug, r.opts...)
	go drainer.Run(workerCtx)
}

// RemoveTenant stops the drainer for a tenant (graceful: existing publishes
// complete). Safe to call if no drainer exists (no-op).
func (r *DrainerRegistry) RemoveTenant(slug string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cancel, ok := r.workers[slug]; ok {
		cancel()
		delete(r.workers, slug)
		slog.Info("outbox: drainer stopped for tenant", "tenant", slug)
	}
}

// ActiveTenants returns the slugs of all tenants with running drainers.
func (r *DrainerRegistry) ActiveTenants() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	slugs := make([]string, 0, len(r.workers))
	for s := range r.workers {
		slugs = append(slugs, s)
	}
	return slugs
}

func (r *DrainerRegistry) stopAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for slug, cancel := range r.workers {
		cancel()
		delete(r.workers, slug)
	}
}
