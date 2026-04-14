// Package service implements the platform business logic.
// The platform service operates on the Platform DB (not tenant DBs).
// It manages tenants, modules, feature flags, and global configuration.
// Only accessible by platform admins (Saldivia team).
package service

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"

	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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

// ListTenants returns tenants (summary view, paginated).
func (p *Platform) ListTenants(ctx context.Context, limit, offset int32) ([]db.ListTenantsRow, error) {
	tenants, err := p.queries.ListTenants(ctx, db.ListTenantsParams{
		Limit:  limit,
		Offset: offset,
	})
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
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	TenantID   string  `json:"tenant_id,omitempty"`
	Enabled    bool    `json:"enabled"`
	RolloutPct int     `json:"rollout_pct"`
	UpdatedAt  string  `json:"updated_at,omitempty"`
	UpdatedBy  *string `json:"updated_by,omitempty"`
}

// CreateFlagParams holds parameters for creating a feature flag.
type CreateFlagParams struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	TenantID    *string `json:"tenant_id,omitempty"`
	RolloutPct  int     `json:"rollout_pct"`
}

// UpdateFlagParams holds parameters for updating a feature flag.
type UpdateFlagParams struct {
	Enabled    *bool   `json:"enabled,omitempty"`
	RolloutPct *int    `json:"rollout_pct,omitempty"`
	UpdatedBy  string  `json:"-"`
}

// ListFeatureFlags returns all feature flags (global + tenant-specific).
func (p *Platform) ListFeatureFlags(ctx context.Context) ([]FeatureFlag, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT id, name, tenant_id, enabled, rollout_pct, updated_at, updated_by
		 FROM feature_flags ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list feature flags: %w", err)
	}
	defer rows.Close()

	var flags []FeatureFlag
	for rows.Next() {
		var f FeatureFlag
		var tenantID, updatedBy *string
		var updatedAt pgtype.Timestamptz
		if err := rows.Scan(&f.ID, &f.Name, &tenantID, &f.Enabled, &f.RolloutPct, &updatedAt, &updatedBy); err != nil {
			return nil, fmt.Errorf("scan feature flag: %w", err)
		}
		if tenantID != nil {
			f.TenantID = *tenantID
		}
		if updatedAt.Valid {
			f.UpdatedAt = updatedAt.Time.Format("2006-01-02T15:04:05Z")
		}
		f.UpdatedBy = updatedBy
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

// CreateFeatureFlag creates a new feature flag (default enabled=false).
func (p *Platform) CreateFeatureFlag(ctx context.Context, params CreateFlagParams, createdBy string) (FeatureFlag, error) {
	if params.RolloutPct < 0 || params.RolloutPct > 100 {
		return FeatureFlag{}, fmt.Errorf("rollout_pct must be 0-100")
	}
	var f FeatureFlag
	err := p.pool.QueryRow(ctx,
		`INSERT INTO feature_flags (id, name, description, tenant_id, enabled, rollout_pct, updated_by, updated_at)
		 VALUES ($1, $2, $3, $4, false, $5, $6, now())
		 RETURNING id, name, tenant_id, enabled, rollout_pct`,
		params.ID, params.Name, params.Description, params.TenantID, params.RolloutPct, createdBy,
	).Scan(&f.ID, &f.Name, &f.TenantID, &f.Enabled, &f.RolloutPct)
	if err != nil {
		return FeatureFlag{}, fmt.Errorf("create feature flag: %w", err)
	}
	p.auditor.Write(ctx, audit.Entry{
		UserID: createdBy, Action: "flag.created", Resource: f.ID,
	})
	p.publishLifecycleEvent("platform", "flag.created", map[string]any{"id": f.ID, "name": f.Name})
	return f, nil
}

// UpdateFeatureFlag updates a feature flag's enabled state and/or rollout percentage.
func (p *Platform) UpdateFeatureFlag(ctx context.Context, id string, params UpdateFlagParams) error {
	if params.RolloutPct != nil && (*params.RolloutPct < 0 || *params.RolloutPct > 100) {
		return fmt.Errorf("rollout_pct must be 0-100")
	}

	var setClauses []string
	var args []any
	argN := 1

	if params.Enabled != nil {
		setClauses = append(setClauses, fmt.Sprintf("enabled = $%d", argN))
		args = append(args, *params.Enabled)
		argN++
	}
	if params.RolloutPct != nil {
		setClauses = append(setClauses, fmt.Sprintf("rollout_pct = $%d", argN))
		args = append(args, *params.RolloutPct)
		argN++
	}
	setClauses = append(setClauses, fmt.Sprintf("updated_by = $%d", argN))
	args = append(args, params.UpdatedBy)
	argN++
	setClauses = append(setClauses, "updated_at = now()")

	args = append(args, id)
	query := fmt.Sprintf("UPDATE feature_flags SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "), argN)

	result, err := p.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update feature flag: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrFlagNotFound
	}
	p.auditor.Write(ctx, audit.Entry{
		UserID: params.UpdatedBy, Action: "flag.updated", Resource: id,
		Details: map[string]any{"enabled": params.Enabled, "rollout_pct": params.RolloutPct},
	})
	p.publishLifecycleEvent("platform", "flag.updated", map[string]any{"id": id})
	return nil
}

// ToggleFeatureFlag enables or disables a feature flag.
func (p *Platform) ToggleFeatureFlag(ctx context.Context, id string, enabled bool) error {
	result, err := p.pool.Exec(ctx,
		`UPDATE feature_flags SET enabled = $2, updated_at = now() WHERE id = $1`,
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
	p.publishLifecycleEvent("platform", "flag.updated", map[string]any{"id": id, "enabled": enabled})
	return nil
}

// KillFlag immediately disables a feature flag (kill switch).
func (p *Platform) KillFlag(ctx context.Context, id string, killedBy string) error {
	result, err := p.pool.Exec(ctx,
		`UPDATE feature_flags SET enabled = false, updated_by = $2, updated_at = now() WHERE id = $1`,
		id, killedBy,
	)
	if err != nil {
		return fmt.Errorf("kill feature flag: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrFlagNotFound
	}
	p.auditor.Write(ctx, audit.Entry{
		UserID: killedBy, Action: "flag.killed", Resource: id,
	})
	p.publishLifecycleEvent("platform", "flag.killed", map[string]any{"id": id})
	return nil
}

// EvaluateFlags evaluates all flags for a given tenant and user.
// Returns a map of flag_name → bool. Uses deterministic hashing for rollout.
func (p *Platform) EvaluateFlags(ctx context.Context, tenantID, userID string) (map[string]bool, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT id, name, tenant_id, enabled, rollout_pct FROM feature_flags
		 WHERE tenant_id IS NULL OR tenant_id = $1
		 ORDER BY name`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("evaluate flags: %w", err)
	}
	defer rows.Close()

	result := make(map[string]bool)
	for rows.Next() {
		var id, name string
		var tid *string
		var enabled bool
		var rolloutPct int
		if err := rows.Scan(&id, &name, &tid, &enabled, &rolloutPct); err != nil {
			return nil, fmt.Errorf("scan flag: %w", err)
		}
		result[name] = enabled && evaluateRollout(id, userID, rolloutPct)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate flags: %w", err)
	}
	return result, nil
}

// evaluateRollout deterministically decides if a user is in the rollout bucket.
func evaluateRollout(flagID, userID string, rolloutPct int) bool {
	if rolloutPct >= 100 {
		return true
	}
	if rolloutPct <= 0 {
		return false
	}
	h := fnv.New32a()
	h.Write([]byte(flagID + ":" + userID))
	return int(h.Sum32()%100) < rolloutPct
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

// ── Deploy Log ─────────────────────────────────────────────────────────

// DeployRecord is a safe representation of a deploy log entry.
type DeployRecord struct {
	ID          string  `json:"id"`
	Service     string  `json:"service"`
	VersionFrom string  `json:"version_from"`
	VersionTo   string  `json:"version_to"`
	Status      string  `json:"status"`
	DeployedBy  string  `json:"deployed_by"`
	StartedAt   string  `json:"started_at"`
	FinishedAt  *string `json:"finished_at,omitempty"`
	Notes       string  `json:"notes,omitempty"`
}

func deployLogToRecord(d db.DeployLog) DeployRecord {
	rec := DeployRecord{
		ID:          d.ID,
		Service:     d.Service,
		VersionFrom: d.VersionFrom,
		VersionTo:   d.VersionTo,
		Status:      d.Status,
		DeployedBy:  d.DeployedBy,
	}
	if d.StartedAt.Valid {
		rec.StartedAt = d.StartedAt.Time.Format("2006-01-02T15:04:05Z")
	}
	if d.FinishedAt.Valid {
		s := d.FinishedAt.Time.Format("2006-01-02T15:04:05Z")
		rec.FinishedAt = &s
	}
	if d.Notes.Valid {
		rec.Notes = d.Notes.String
	}
	return rec
}

// RecordDeploy inserts a deploy log entry and publishes a NATS lifecycle event.
func (p *Platform) RecordDeploy(ctx context.Context, arg db.InsertDeployLogParams) (DeployRecord, error) {
	row, err := p.queries.InsertDeployLog(ctx, arg)
	if err != nil {
		return DeployRecord{}, fmt.Errorf("record deploy: %w", err)
	}
	rec := deployLogToRecord(row)
	p.publishLifecycleEvent("platform", "deploy.created", rec)
	p.auditor.Write(ctx, audit.Entry{
		UserID: arg.DeployedBy, Action: "deploy.created", Resource: rec.ID,
		Details: map[string]any{"service": arg.Service, "from": arg.VersionFrom, "to": arg.VersionTo},
	})
	return rec, nil
}

// ListDeploys returns recent deploy log entries, newest first.
func (p *Platform) ListDeploys(ctx context.Context, limit, offset int32) ([]DeployRecord, error) {
	rows, err := p.queries.ListDeployLogs(ctx, db.ListDeployLogsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list deploys: %w", err)
	}
	records := make([]DeployRecord, 0, len(rows))
	for _, r := range rows {
		records = append(records, deployLogToRecord(r))
	}
	return records, nil
}

// isDuplicateKey checks for unique constraint violation (PG error code 23505).
func isDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
