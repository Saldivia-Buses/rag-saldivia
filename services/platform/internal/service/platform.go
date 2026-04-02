// Package service implements the platform business logic.
// The platform service operates on the Platform DB (not tenant DBs).
// It manages tenants, modules, feature flags, and global configuration.
// Only accessible by platform admins (Saldivia team).
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Camionerou/rag-saldivia/services/platform/db"
)

var (
	ErrTenantNotFound = errors.New("tenant not found")
	ErrModuleNotFound = errors.New("module not found")
	ErrSlugTaken      = errors.New("tenant slug already taken")
)

// Platform handles platform operations.
type Platform struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// New creates a platform service.
func New(pool *pgxpool.Pool) *Platform {
	return &Platform{
		pool:    pool,
		queries: db.New(pool),
	}
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

// GetTenant returns a tenant by slug (full detail).
func (p *Platform) GetTenant(ctx context.Context, slug string) (db.Tenant, error) {
	tenant, err := p.queries.GetTenantBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Tenant{}, ErrTenantNotFound
		}
		return db.Tenant{}, fmt.Errorf("get tenant: %w", err)
	}
	return tenant, nil
}

// CreateTenant creates a new tenant.
func (p *Platform) CreateTenant(ctx context.Context, arg db.CreateTenantParams) (db.Tenant, error) {
	tenant, err := p.queries.CreateTenant(ctx, arg)
	if err != nil {
		if isDuplicateKey(err) {
			return db.Tenant{}, ErrSlugTaken
		}
		return db.Tenant{}, fmt.Errorf("create tenant: %w", err)
	}
	return tenant, nil
}

// UpdateTenant updates a tenant's name, plan, and settings.
func (p *Platform) UpdateTenant(ctx context.Context, arg db.UpdateTenantParams) error {
	if err := p.queries.UpdateTenant(ctx, arg); err != nil {
		return fmt.Errorf("update tenant: %w", err)
	}
	return nil
}

// DisableTenant disables a tenant (soft delete).
func (p *Platform) DisableTenant(ctx context.Context, id string) error {
	if err := p.queries.DisableTenant(ctx, id); err != nil {
		return fmt.Errorf("disable tenant: %w", err)
	}
	return nil
}

// EnableTenant re-enables a disabled tenant.
func (p *Platform) EnableTenant(ctx context.Context, id string) error {
	if err := p.queries.EnableTenant(ctx, id); err != nil {
		return fmt.Errorf("enable tenant: %w", err)
	}
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
		return errors.New("feature flag not found")
	}
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
			return ConfigEntry{}, errors.New("config key not found")
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
	return nil
}

// isDuplicateKey checks for unique constraint violation.
func isDuplicateKey(err error) bool {
	return err != nil && (errors.Is(err, pgx.ErrNoRows) == false) &&
		(fmt.Sprintf("%v", err) != "" && contains(err.Error(), "duplicate key") || contains(err.Error(), "23505"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
