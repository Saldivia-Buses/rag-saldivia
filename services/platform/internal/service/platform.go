// Package service implements the platform business logic.
// The platform service operates on the Platform DB (not tenant DBs).
// It manages tenants, modules, feature flags, and global configuration.
// Only accessible by platform admins (Saldivia team).
package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	"github.com/Camionerou/rag-saldivia/services/platform/db"
)

var slugRegex = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}$`)

var (
	ErrTenantNotFound = errors.New("tenant not found")
	ErrModuleNotFound = errors.New("module not found")
	ErrSlugTaken      = errors.New("tenant slug already taken")
	ErrInvalidSlug    = errors.New("slug must be lowercase alphanumeric with hyphens, 2-63 chars")
	ErrFlagNotFound   = errors.New("feature flag not found")
	ErrConfigNotFound = errors.New("config key not found")
)

// Platform handles platform operations.
type Platform struct {
	pool      *pgxpool.Pool
	queries   *db.Queries
	publisher *natspub.Publisher
	auditor   *audit.Writer
}

// New creates a platform service.
func New(pool *pgxpool.Pool, publisher *natspub.Publisher) *Platform {
	return &Platform{
		pool:      pool,
		queries:   db.New(pool),
		publisher: publisher,
		auditor:   audit.NewWriter(pool),
	}
}

// publishLifecycleEvent emits a NATS event for config/tenant changes.
// Other services can react without polling or restarting.
// Event type dots are replaced with underscores to keep NATS subjects
// at 4 segments (tenant.{slug}.notify.{type}) matching the permission
// grant tenant.*.notify.* in nats-server.conf.
func (p *Platform) publishLifecycleEvent(tenantSlug, eventType string, data any) {
	if p.publisher == nil || tenantSlug == "" {
		return
	}
	safeType := "platform_" + strings.ReplaceAll(eventType, ".", "_")
	if err := p.publisher.Notify(tenantSlug, map[string]any{
		"type": safeType,
		"data": data,
	}); err != nil {
		slog.Warn("publish lifecycle event failed", "event", eventType, "error", err)
	}
}

// TenantDetail is a safe representation of a tenant (no connection strings).
type TenantDetail struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	PlanID    string `json:"plan_id"`
	Enabled   bool   `json:"enabled"`
	LogoURL   string `json:"logo_url,omitempty"`
	Domain    string `json:"domain,omitempty"`
	Settings  []byte `json:"settings"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func tenantToDetail(t db.Tenant) TenantDetail {
	d := TenantDetail{
		ID:        t.ID,
		Slug:      t.Slug,
		Name:      t.Name,
		PlanID:    t.PlanID,
		Enabled:   t.Enabled,
		Settings:  t.Settings,
		CreatedAt: t.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: t.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if t.LogoUrl.Valid {
		d.LogoURL = t.LogoUrl.String
	}
	if t.Domain.Valid {
		d.Domain = t.Domain.String
	}
	return d
}

func slugRowToDetail(t db.GetTenantBySlugRow) TenantDetail {
	d := TenantDetail{
		ID:        t.ID,
		Slug:      t.Slug,
		Name:      t.Name,
		PlanID:    t.PlanID,
		Enabled:   t.Enabled,
		Settings:  t.Settings,
		CreatedAt: t.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: t.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if t.LogoUrl.Valid {
		d.LogoURL = t.LogoUrl.String
	}
	if t.Domain.Valid {
		d.Domain = t.Domain.String
	}
	return d
}

func createRowToDetail(t db.CreateTenantRow) TenantDetail {
	d := TenantDetail{
		ID:        t.ID,
		Slug:      t.Slug,
		Name:      t.Name,
		PlanID:    t.PlanID,
		Enabled:   t.Enabled,
		Settings:  t.Settings,
		CreatedAt: t.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: t.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if t.LogoUrl.Valid {
		d.LogoURL = t.LogoUrl.String
	}
	if t.Domain.Valid {
		d.Domain = t.Domain.String
	}
	return d
}

// ── Tenants ─────────────────────────────────────────────────────────────

// ListTenants returns all tenants (summary view).
func (p *Platform) ListTenants(ctx context.Context) ([]db.ListTenantsRow, error) {
	tenants, err := p.queries.ListTenants(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tenants: %w", err)
	}
	return tenants, nil
}

// GetTenant returns a tenant by slug (safe view without connection strings).
func (p *Platform) GetTenant(ctx context.Context, slug string) (TenantDetail, error) {
	tenant, err := p.queries.GetTenantBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TenantDetail{}, ErrTenantNotFound
		}
		return TenantDetail{}, fmt.Errorf("get tenant: %w", err)
	}
	return slugRowToDetail(tenant), nil
}

// CreateTenant creates a new tenant.
func (p *Platform) CreateTenant(ctx context.Context, arg db.CreateTenantParams) (TenantDetail, error) {
	if !slugRegex.MatchString(arg.Slug) {
		return TenantDetail{}, ErrInvalidSlug
	}

	tenant, err := p.queries.CreateTenant(ctx, arg)
	if err != nil {
		if isDuplicateKey(err) {
			return TenantDetail{}, ErrSlugTaken
		}
		return TenantDetail{}, fmt.Errorf("create tenant: %w", err)
	}
	detail := createRowToDetail(tenant)
	p.publishLifecycleEvent(arg.Slug, "tenant.created", detail)
	p.auditor.Write(ctx, audit.Entry{
		Action: "tenant.created", Resource: detail.ID,
		Details: map[string]any{"slug": arg.Slug, "name": arg.Name, "plan": arg.PlanID},
	})
	return detail, nil
}

// UpdateTenant updates a tenant's name, plan, and settings.
func (p *Platform) UpdateTenant(ctx context.Context, arg db.UpdateTenantParams) error {
	if err := p.queries.UpdateTenant(ctx, arg); err != nil {
		return fmt.Errorf("update tenant: %w", err)
	}
	p.auditor.Write(ctx, audit.Entry{
		Action: "tenant.updated", Resource: arg.ID,
		Details: map[string]any{"name": arg.Name, "plan": arg.PlanID},
	})
	return nil
}

// DisableTenant disables a tenant (soft delete).
func (p *Platform) DisableTenant(ctx context.Context, id string) error {
	if err := p.queries.DisableTenant(ctx, id); err != nil {
		return fmt.Errorf("disable tenant: %w", err)
	}
	p.auditor.Write(ctx, audit.Entry{Action: "tenant.disabled", Resource: id})
	return nil
}

// EnableTenant re-enables a disabled tenant.
func (p *Platform) EnableTenant(ctx context.Context, id string) error {
	if err := p.queries.EnableTenant(ctx, id); err != nil {
		return fmt.Errorf("enable tenant: %w", err)
	}
	p.auditor.Write(ctx, audit.Entry{Action: "tenant.enabled", Resource: id})
	return nil
}

// ── Modules ─────────────────────────────────────────────────────────────

// ListModules returns all available modules.
func (p *Platform) ListModules(ctx context.Context) ([]db.Module, error) {
	modules, err := p.queries.ListModules(ctx)
	if err != nil {
		return nil, fmt.Errorf("list modules: %w", err)
	}
	return modules, nil
}

// GetTenantModules returns modules enabled for a specific tenant.
func (p *Platform) GetTenantModules(ctx context.Context, tenantID string) ([]db.GetEnabledModulesForTenantRow, error) {
	modules, err := p.queries.GetEnabledModulesForTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get tenant modules: %w", err)
	}
	return modules, nil
}

// EnableModule enables a module for a tenant.
func (p *Platform) EnableModule(ctx context.Context, arg db.EnableModuleForTenantParams) error {
	if err := p.queries.EnableModuleForTenant(ctx, arg); err != nil {
		return fmt.Errorf("enable module: %w", err)
	}
	p.auditor.Write(ctx, audit.Entry{
		Action: "module.enabled", Resource: arg.ModuleID,
		Details: map[string]any{"tenant_id": arg.TenantID, "enabled_by": arg.EnabledBy},
	})
	return nil
}

// DisableModule disables a module for a tenant.
func (p *Platform) DisableModule(ctx context.Context, tenantID, moduleID string) error {
	if err := p.queries.DisableModuleForTenant(ctx, db.DisableModuleForTenantParams{
		TenantID: tenantID,
		ModuleID: moduleID,
	}); err != nil {
		return fmt.Errorf("disable module: %w", err)
	}
	p.auditor.Write(ctx, audit.Entry{
		Action: "module.disabled", Resource: moduleID,
		Details: map[string]any{"tenant_id": tenantID},
	})
	return nil
}

// ── Feature Flags ───────────────────────────────────────────────────────

// FeatureFlag represents a feature flag.
type FeatureFlag struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	TenantID string `json:"tenant_id,omitempty"`
	Enabled  bool   `json:"enabled"`
}

// ListFeatureFlags returns all feature flags (global + tenant-specific).
func (p *Platform) ListFeatureFlags(ctx context.Context) ([]FeatureFlag, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT id, name, tenant_id, enabled FROM feature_flags ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list feature flags: %w", err)
	}
	defer rows.Close()

	var flags []FeatureFlag
	for rows.Next() {
		var f FeatureFlag
		var tenantID *string
		if err := rows.Scan(&f.ID, &f.Name, &tenantID, &f.Enabled); err != nil {
			return nil, fmt.Errorf("scan feature flag: %w", err)
		}
		if tenantID != nil {
			f.TenantID = *tenantID
		}
		flags = append(flags, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate feature flags: %w", err)
	}
	if flags == nil {
		flags = []FeatureFlag{}
	}
	return flags, nil
}

// ToggleFeatureFlag enables or disables a feature flag.
func (p *Platform) ToggleFeatureFlag(ctx context.Context, id string, enabled bool) error {
	result, err := p.pool.Exec(ctx,
		`UPDATE feature_flags SET enabled = $2 WHERE id = $1`,
		id, enabled,
	)
	if err != nil {
		return fmt.Errorf("toggle feature flag: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrFlagNotFound
	}
	p.auditor.Write(ctx, audit.Entry{
		Action: "flag.toggled", Resource: id,
		Details: map[string]any{"enabled": enabled},
	})
	return nil
}

// ── Global Config ───────────────────────────────────────────────────────

// ConfigEntry represents a global configuration entry.
type ConfigEntry struct {
	Key       string `json:"key"`
	Value     []byte `json:"value"`
	UpdatedBy string `json:"updated_by"`
}

// GetConfig returns a global config value.
func (p *Platform) GetConfig(ctx context.Context, key string) (ConfigEntry, error) {
	var c ConfigEntry
	err := p.pool.QueryRow(ctx,
		`SELECT key, value, updated_by FROM global_config WHERE key = $1`,
		key,
	).Scan(&c.Key, &c.Value, &c.UpdatedBy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ConfigEntry{}, ErrConfigNotFound
		}
		return ConfigEntry{}, fmt.Errorf("get config: %w", err)
	}
	return c, nil
}

// SetConfig upserts a global config value.
func (p *Platform) SetConfig(ctx context.Context, key string, value []byte, updatedBy string) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO global_config (key, value, updated_by, updated_at)
		 VALUES ($1, $2, $3, now())
		 ON CONFLICT (key) DO UPDATE SET value = $2, updated_by = $3, updated_at = now()`,
		key, value, updatedBy,
	)
	if err != nil {
		return fmt.Errorf("set config: %w", err)
	}
	p.auditor.Write(ctx, audit.Entry{
		UserID: updatedBy, Action: "config.updated", Resource: key,
	})
	return nil
}

// isDuplicateKey checks for unique constraint violation (PG error code 23505).
func isDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
