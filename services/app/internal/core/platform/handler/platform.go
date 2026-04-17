// Package handler implements HTTP handlers for the platform service.
// All endpoints require platform admin role verified via JWT.
package handler

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/httperr"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/pkg/tenant"
	"github.com/Camionerou/rag-saldivia/services/app/internal/core/platform/db"
	"github.com/Camionerou/rag-saldivia/services/app/internal/core/platform/service"
)

// PlatformService defines the operations the handler needs from the service layer.
type PlatformService interface {
	ListTenants(ctx context.Context, limit, offset int32) ([]db.ListTenantsRow, error)
	GetTenant(ctx context.Context, slug string) (service.TenantDetail, error)
	GetTenantByID(ctx context.Context, id string) (service.TenantDetail, error)
	CreateTenant(ctx context.Context, arg db.CreateTenantParams) (service.TenantDetail, error)
	UpdateTenant(ctx context.Context, arg db.UpdateTenantParams) error
	DisableTenant(ctx context.Context, id string) error
	EnableTenant(ctx context.Context, id string) error
	ListModules(ctx context.Context) ([]db.Module, error)
	GetTenantModules(ctx context.Context, tenantID string) ([]db.GetEnabledModulesForTenantRow, error)
	EnableModule(ctx context.Context, arg db.EnableModuleForTenantParams) error
	DisableModule(ctx context.Context, tenantID, moduleID string) error
	ListFeatureFlags(ctx context.Context) ([]service.FeatureFlag, error)
	CreateFeatureFlag(ctx context.Context, params service.CreateFlagParams, createdBy string) (service.FeatureFlag, error)
	UpdateFeatureFlag(ctx context.Context, id string, params service.UpdateFlagParams) error
	ToggleFeatureFlag(ctx context.Context, id string, enabled bool) error
	KillFlag(ctx context.Context, id string, killedBy string) error
	EvaluateFlags(ctx context.Context, tenantID, userID string) (map[string]bool, error)
	GetConfig(ctx context.Context, key string) (service.ConfigEntry, error)
	SetConfig(ctx context.Context, key string, value []byte, updatedBy string) error
	RecordDeploy(ctx context.Context, arg db.InsertDeployLogParams) (service.DeployRecord, error)
	ListDeploys(ctx context.Context, limit, offset int32) ([]service.DeployRecord, error)
}

// Platform handles HTTP requests for platform operations.
type Platform struct {
	svc          PlatformService
	publicKey    ed25519.PublicKey
	platformSlug string // tenant slug that identifies platform admins
	blacklist    *security.TokenBlacklist
}

// NewPlatform creates platform HTTP handlers.
func NewPlatform(svc PlatformService, publicKey ed25519.PublicKey, platformSlug string, blacklist *security.TokenBlacklist) *Platform {
	return &Platform{svc: svc, publicKey: publicKey, platformSlug: platformSlug, blacklist: blacklist}
}

// Routes returns a chi router with all platform routes.
func (h *Platform) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(h.requirePlatformAdmin)

	// Tenants
	r.Route("/tenants", func(r chi.Router) {
		r.Get("/", h.ListTenants)
		r.Post("/", h.CreateTenant)
		r.Get("/by-id/{tenantID}", h.GetTenantByID)
		r.Get("/{slug}", h.GetTenant)
		r.Put("/{tenantID}", h.UpdateTenant)
		r.Post("/{tenantID}/disable", h.DisableTenant)
		r.Post("/{tenantID}/enable", h.EnableTenant)

		// Tenant modules
		r.Get("/{tenantID}/modules", h.GetTenantModules)
		r.Post("/{tenantID}/modules", h.EnableModule)
		r.Delete("/{tenantID}/modules/{moduleID}", h.DisableModule)
	})

	// Modules catalog
	r.Get("/modules", h.ListModules)

	// Feature flags (admin CRUD)
	r.Get("/flags", h.ListFeatureFlags)
	r.Post("/flags", h.CreateFeatureFlag)
	r.Put("/flags/{flagID}", h.UpdateFeatureFlag)
	r.Patch("/flags/{flagID}", h.ToggleFeatureFlag)
	r.Delete("/flags/{flagID}/kill", h.KillFlag)

	// Global config
	r.Get("/config/{key}", h.GetConfig)
	r.Put("/config/{key}", h.SetConfig)

	// Deploy log
	r.Post("/deploys", h.RecordDeploy)
	r.Get("/deploys", h.ListDeploys)

	return r
}

// FlagsRoutes returns a chi router for flag evaluation (standard auth, not platform admin).
func (h *Platform) FlagsRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(sdamw.Auth(h.publicKey))
	r.Get("/evaluate", h.EvaluateFlags)
	return r
}

// ── Tenants ─────────────────────────────────────────────────────────────

// ListTenants handles GET /v1/platform/tenants (paginated)
func (h *Platform) ListTenants(w http.ResponseWriter, r *http.Request) {
	pg := pagination.Parse(r)
	tenants, err := h.svc.ListTenants(r.Context(), int32(pg.Limit()), int32(pg.Offset()))
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, tenants)
}

// GetTenant handles GET /v1/platform/tenants/{slug}
func (h *Platform) GetTenant(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	tenant, err := h.svc.GetTenant(r.Context(), slug)
	if err != nil {
		if errors.Is(err, service.ErrTenantNotFound) {
			httperr.WriteError(w, r, httperr.NotFound("tenant"))
			return
		}
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, tenant)
}

// GetTenantByID handles GET /v1/platform/tenants/by-id/{tenantID}
func (h *Platform) GetTenantByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "tenantID")
	tenant, err := h.svc.GetTenantByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTenantNotFound) {
			httperr.WriteError(w, r, httperr.NotFound("tenant"))
			return
		}
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, tenant)
}

type createTenantRequest struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	PlanID      string `json:"plan_id"`
	PostgresURL string `json:"postgres_url"`
	RedisURL    string `json:"redis_url"`
}

// CreateTenant handles POST /v1/platform/tenants
func (h *Platform) CreateTenant(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}

	if req.Slug == "" || req.Name == "" || req.PlanID == "" || req.PostgresURL == "" || req.RedisURL == "" {
		httperr.WriteError(w, r, httperr.InvalidInput("slug, name, plan_id, postgres_url, and redis_url are required"))
		return
	}

	tenant, err := h.svc.CreateTenant(r.Context(), db.CreateTenantParams{
		Slug:        req.Slug,
		Name:        req.Name,
		PlanID:      req.PlanID,
		PostgresUrl: req.PostgresURL,
		RedisUrl:    req.RedisURL,
		Settings:    []byte("{}"),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidSlug) {
			httperr.WriteError(w, r, httperr.InvalidInput("slug must be lowercase alphanumeric with hyphens, 2-63 chars"))
			return
		}
		if errors.Is(err, service.ErrSlugTaken) {
			httperr.WriteError(w, r, httperr.Conflict("tenant slug already taken"))
			return
		}
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusCreated, tenant)
}

type updateTenantRequest struct {
	Name     string          `json:"name"`
	PlanID   string          `json:"plan_id"`
	Settings json.RawMessage `json:"settings"`
}

// UpdateTenant handles PUT /v1/platform/tenants/{tenantID}
func (h *Platform) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	tenantID := chi.URLParam(r, "tenantID")

	var req updateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}

	settings := req.Settings
	if settings == nil {
		settings = json.RawMessage("{}")
	}

	if err := h.svc.UpdateTenant(r.Context(), db.UpdateTenantParams{
		ID:       tenantID,
		Name:     req.Name,
		PlanID:   req.PlanID,
		Settings: settings,
	}); err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DisableTenant handles POST /v1/platform/tenants/{tenantID}/disable
func (h *Platform) DisableTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if err := h.svc.DisableTenant(r.Context(), tenantID); err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// EnableTenant handles POST /v1/platform/tenants/{tenantID}/enable
func (h *Platform) EnableTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if err := h.svc.EnableTenant(r.Context(), tenantID); err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Modules ─────────────────────────────────────────────────────────────

// ListModules handles GET /v1/platform/modules
func (h *Platform) ListModules(w http.ResponseWriter, r *http.Request) {
	modules, err := h.svc.ListModules(r.Context())
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, modules)
}

// GetTenantModules handles GET /v1/platform/tenants/{tenantID}/modules
func (h *Platform) GetTenantModules(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	modules, err := h.svc.GetTenantModules(r.Context(), tenantID)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, modules)
}

type enableModuleRequest struct {
	ModuleID string          `json:"module_id"`
	Config   json.RawMessage `json:"config"`
}

// EnableModule handles POST /v1/platform/tenants/{tenantID}/modules
func (h *Platform) EnableModule(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	tenantID := chi.URLParam(r, "tenantID")
	adminID := r.Header.Get("X-User-ID")

	var req enableModuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}

	if req.ModuleID == "" {
		httperr.WriteError(w, r, httperr.InvalidInput("module_id is required"))
		return
	}

	config := req.Config
	if config == nil {
		config = json.RawMessage("{}")
	}

	if err := h.svc.EnableModule(r.Context(), db.EnableModuleForTenantParams{
		TenantID:  tenantID,
		ModuleID:  req.ModuleID,
		Config:    config,
		EnabledBy: adminID,
	}); err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DisableModule handles DELETE /v1/platform/tenants/{tenantID}/modules/{moduleID}
func (h *Platform) DisableModule(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	moduleID := chi.URLParam(r, "moduleID")

	if err := h.svc.DisableModule(r.Context(), tenantID, moduleID); err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Feature Flags ───────────────────────────────────────────────────────

// ListFeatureFlags handles GET /v1/platform/flags
func (h *Platform) ListFeatureFlags(w http.ResponseWriter, r *http.Request) {
	flags, err := h.svc.ListFeatureFlags(r.Context())
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, flags)
}

type toggleFlagRequest struct {
	Enabled bool `json:"enabled"`
}

// ToggleFeatureFlag handles PATCH /v1/platform/flags/{flagID}
func (h *Platform) ToggleFeatureFlag(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	flagID := chi.URLParam(r, "flagID")

	var req toggleFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}

	if err := h.svc.ToggleFeatureFlag(r.Context(), flagID, req.Enabled); err != nil {
		if errors.Is(err, service.ErrFlagNotFound) {
			httperr.WriteError(w, r, httperr.NotFound("feature flag"))
			return
		}
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type createFlagRequest struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	TenantID    *string `json:"tenant_id,omitempty"`
	RolloutPct  *int    `json:"rollout_pct,omitempty"`
}

// CreateFeatureFlag handles POST /v1/platform/flags
func (h *Platform) CreateFeatureFlag(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req createFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}
	if req.ID == "" || req.Name == "" {
		httperr.WriteError(w, r, httperr.InvalidInput("id and name are required"))
		return
	}

	rollout := 0
	if req.RolloutPct != nil {
		rollout = *req.RolloutPct
		if rollout < 0 || rollout > 100 {
			httperr.WriteError(w, r, httperr.InvalidInput("rollout_pct must be 0-100"))
			return
		}
	}

	createdBy := r.Header.Get("X-User-ID")

	flag, err := h.svc.CreateFeatureFlag(r.Context(), service.CreateFlagParams{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		TenantID:    req.TenantID,
		RolloutPct:  rollout,
	}, createdBy)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusCreated, flag)
}

type updateFlagRequest struct {
	Enabled    *bool `json:"enabled,omitempty"`
	RolloutPct *int  `json:"rollout_pct,omitempty"`
}

// UpdateFeatureFlag handles PUT /v1/platform/flags/{flagID}
func (h *Platform) UpdateFeatureFlag(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	flagID := chi.URLParam(r, "flagID")

	var req updateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}
	if req.Enabled == nil && req.RolloutPct == nil {
		httperr.WriteError(w, r, httperr.InvalidInput("at least one of enabled or rollout_pct is required"))
		return
	}
	if req.RolloutPct != nil && (*req.RolloutPct < 0 || *req.RolloutPct > 100) {
		httperr.WriteError(w, r, httperr.InvalidInput("rollout_pct must be 0-100"))
		return
	}

	updatedBy := r.Header.Get("X-User-ID")

	err := h.svc.UpdateFeatureFlag(r.Context(), flagID, service.UpdateFlagParams{
		Enabled:    req.Enabled,
		RolloutPct: req.RolloutPct,
		UpdatedBy:  updatedBy,
	})
	if err != nil {
		if errors.Is(err, service.ErrFlagNotFound) {
			httperr.WriteError(w, r, httperr.NotFound("feature flag"))
			return
		}
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// KillFlag handles DELETE /v1/platform/flags/{flagID}/kill
func (h *Platform) KillFlag(w http.ResponseWriter, r *http.Request) {
	flagID := chi.URLParam(r, "flagID")

	killedBy := r.Header.Get("X-User-ID")

	err := h.svc.KillFlag(r.Context(), flagID, killedBy)
	if err != nil {
		if errors.Is(err, service.ErrFlagNotFound) {
			httperr.WriteError(w, r, httperr.NotFound("feature flag"))
			return
		}
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// EvaluateFlags handles GET /v1/flags/evaluate
// Identity comes from JWT claims only (DS7 — no query params for identity).
func (h *Platform) EvaluateFlags(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")
	if tenantID == "" || userID == "" {
		httperr.WriteError(w, r, httperr.Unauthorized("missing identity"))
		return
	}

	flags, err := h.svc.EvaluateFlags(r.Context(), tenantID, userID)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"flags": flags})
}

// ── Global Config ───────────────────────────────────────────────────────

// GetConfig handles GET /v1/platform/config/{key}
func (h *Platform) GetConfig(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	config, err := h.svc.GetConfig(r.Context(), key)
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			httperr.WriteError(w, r, httperr.NotFound("config key"))
			return
		}
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, config)
}

type setConfigRequest struct {
	Value json.RawMessage `json:"value"`
}

// SetConfig handles PUT /v1/platform/config/{key}
func (h *Platform) SetConfig(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	key := chi.URLParam(r, "key")
	adminID := r.Header.Get("X-User-ID")

	var req setConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}

	if err := h.svc.SetConfig(r.Context(), key, req.Value, adminID); err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Deploy Log ──────────────────────────────────────────────────────────

type recordDeployRequest struct {
	Service     string `json:"service"`
	VersionFrom string `json:"version_from"`
	VersionTo   string `json:"version_to"`
	Status      string `json:"status"`
	Notes       string `json:"notes,omitempty"`
}

// RecordDeploy handles POST /v1/platform/deploys
func (h *Platform) RecordDeploy(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req recordDeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.WriteError(w, r, httperr.InvalidInput("invalid request body"))
		return
	}

	if req.Service == "" || req.VersionTo == "" {
		httperr.WriteError(w, r, httperr.InvalidInput("service and version_to are required"))
		return
	}

	deployedBy := r.Header.Get("X-User-ID")
	if deployedBy == "" {
		deployedBy = "system"
	}

	status := req.Status
	if status == "" {
		status = "pending"
	}

	switch status {
	case "pending", "success", "failed", "rollback":
		// valid
	default:
		httperr.WriteError(w, r, httperr.InvalidInput("status must be one of: pending, success, failed, rollback"))
		return
	}

	var notes pgtype.Text
	if req.Notes != "" {
		notes = pgtype.Text{String: req.Notes, Valid: true}
	}

	record, err := h.svc.RecordDeploy(r.Context(), db.InsertDeployLogParams{
		Service:     req.Service,
		VersionFrom: req.VersionFrom,
		VersionTo:   req.VersionTo,
		Status:      status,
		DeployedBy:  deployedBy,
		Notes:       notes,
	})
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusCreated, record)
}

// ListDeploys handles GET /v1/platform/deploys
func (h *Platform) ListDeploys(w http.ResponseWriter, r *http.Request) {
	pg := pagination.Parse(r)
	deploys, err := h.svc.ListDeploys(r.Context(), int32(pg.Limit()), int32(pg.Offset()))
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, deploys)
}

// ── Middleware & Helpers ─────────────────────────────────────────────────

// requirePlatformAdmin verifies JWT, checks blacklist, strips spoofed headers,
// and ensures the user has admin role on the platform tenant.
func (h *Platform) requirePlatformAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// H2: strip spoofed identity headers before processing
		r.Header.Del("X-User-ID")
		r.Header.Del("X-User-Email")
		r.Header.Del("X-User-Role")
		r.Header.Del("X-Tenant-ID")
		r.Header.Del("X-Tenant-Slug")

		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			httperr.WriteError(w, r, httperr.Unauthorized("missing authorization"))
			return
		}

		claims, err := sdajwt.Verify(h.publicKey, strings.TrimPrefix(auth, "Bearer "))
		if err != nil {
			httperr.WriteError(w, r, httperr.Unauthorized("invalid token"))
			return
		}

		// Check blacklist (revoked tokens) — fail-closed for admin endpoints
		if h.blacklist != nil && claims.ID != "" {
			revoked, err := h.blacklist.IsRevoked(r.Context(), claims.ID)
			if err != nil {
				httperr.WriteError(w, r, httperr.Wrap(err, httperr.CodeInternal, "auth check unavailable", 503))
				return
			}
			if revoked {
				httperr.WriteError(w, r, httperr.Unauthorized("token revoked"))
				return
			}
		}

		if claims.Role != "admin" || claims.Slug != h.platformSlug {
			httperr.WriteError(w, r, httperr.Forbidden("platform admin access required"))
			return
		}

		// Propagate identity via headers for downstream use
		r.Header.Set("X-User-ID", claims.UserID)
		r.Header.Set("X-User-Email", claims.Email)
		r.Header.Set("X-User-Role", claims.Role)
		r.Header.Set("X-Tenant-ID", claims.TenantID)
		r.Header.Set("X-Tenant-Slug", claims.Slug)

		// Set context values to align with pkg/middleware.AuthWithConfig
		ctx := tenant.WithInfo(r.Context(), tenant.Info{
			ID:   claims.TenantID,
			Slug: claims.Slug,
		})
		ctx = sdamw.WithRole(ctx, claims.Role)
		ctx = sdamw.WithUserID(ctx, claims.UserID)
		ctx = sdamw.WithUserEmail(ctx, claims.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

