// Package handler implements HTTP handlers for the platform service.
// All endpoints require platform admin role verified via JWT.
package handler

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/services/platform/db"
	"github.com/Camionerou/rag-saldivia/services/platform/internal/service"
)

// PlatformService defines the operations the handler needs from the service layer.
type PlatformService interface {
	ListTenants(ctx context.Context) ([]db.ListTenantsRow, error)
	GetTenant(ctx context.Context, slug string) (service.TenantDetail, error)
	CreateTenant(ctx context.Context, arg db.CreateTenantParams) (service.TenantDetail, error)
	UpdateTenant(ctx context.Context, arg db.UpdateTenantParams) error
	DisableTenant(ctx context.Context, id string) error
	EnableTenant(ctx context.Context, id string) error
	ListModules(ctx context.Context) ([]db.Module, error)
	GetTenantModules(ctx context.Context, tenantID string) ([]db.GetEnabledModulesForTenantRow, error)
	EnableModule(ctx context.Context, arg db.EnableModuleForTenantParams) error
	DisableModule(ctx context.Context, tenantID, moduleID string) error
	ListFeatureFlags(ctx context.Context) ([]service.FeatureFlag, error)
	ToggleFeatureFlag(ctx context.Context, id string, enabled bool) error
	GetConfig(ctx context.Context, key string) (service.ConfigEntry, error)
	SetConfig(ctx context.Context, key string, value []byte, updatedBy string) error
}

// Platform handles HTTP requests for platform operations.
type Platform struct {
	svc          PlatformService
	publicKey    ed25519.PublicKey
	platformSlug string // tenant slug that identifies platform admins
}

// NewPlatform creates platform HTTP handlers.
func NewPlatform(svc PlatformService, publicKey ed25519.PublicKey, platformSlug string) *Platform {
	return &Platform{svc: svc, publicKey: publicKey, platformSlug: platformSlug}
}

// Routes returns a chi router with all platform routes.
func (h *Platform) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(h.requirePlatformAdmin)

	// Tenants
	r.Route("/tenants", func(r chi.Router) {
		r.Get("/", h.ListTenants)
		r.Post("/", h.CreateTenant)
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

	// Feature flags
	r.Get("/flags", h.ListFeatureFlags)
	r.Patch("/flags/{flagID}", h.ToggleFeatureFlag)

	// Global config
	r.Get("/config/{key}", h.GetConfig)
	r.Put("/config/{key}", h.SetConfig)

	return r
}

// ── Tenants ─────────────────────────────────────────────────────────────

// ListTenants handles GET /v1/platform/tenants
func (h *Platform) ListTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.svc.ListTenants(r.Context())
	if err != nil {
		serverError(w, r, err)
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
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "tenant not found"})
			return
		}
		serverError(w, r, err)
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Slug == "" || req.Name == "" || req.PlanID == "" || req.PostgresURL == "" || req.RedisURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "slug, name, plan_id, postgres_url, and redis_url are required"})
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
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "slug must be lowercase alphanumeric with hyphens, 2-63 chars"})
			return
		}
		if errors.Is(err, service.ErrSlugTaken) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "tenant slug already taken"})
			return
		}
		serverError(w, r, err)
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
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
		serverError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DisableTenant handles POST /v1/platform/tenants/{tenantID}/disable
func (h *Platform) DisableTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if err := h.svc.DisableTenant(r.Context(), tenantID); err != nil {
		serverError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// EnableTenant handles POST /v1/platform/tenants/{tenantID}/enable
func (h *Platform) EnableTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if err := h.svc.EnableTenant(r.Context(), tenantID); err != nil {
		serverError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Modules ─────────────────────────────────────────────────────────────

// ListModules handles GET /v1/platform/modules
func (h *Platform) ListModules(w http.ResponseWriter, r *http.Request) {
	modules, err := h.svc.ListModules(r.Context())
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, modules)
}

// GetTenantModules handles GET /v1/platform/tenants/{tenantID}/modules
func (h *Platform) GetTenantModules(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	modules, err := h.svc.GetTenantModules(r.Context(), tenantID)
	if err != nil {
		serverError(w, r, err)
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.ModuleID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "module_id is required"})
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
		serverError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DisableModule handles DELETE /v1/platform/tenants/{tenantID}/modules/{moduleID}
func (h *Platform) DisableModule(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	moduleID := chi.URLParam(r, "moduleID")

	if err := h.svc.DisableModule(r.Context(), tenantID, moduleID); err != nil {
		serverError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Feature Flags ───────────────────────────────────────────────────────

// ListFeatureFlags handles GET /v1/platform/flags
func (h *Platform) ListFeatureFlags(w http.ResponseWriter, r *http.Request) {
	flags, err := h.svc.ListFeatureFlags(r.Context())
	if err != nil {
		serverError(w, r, err)
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.svc.ToggleFeatureFlag(r.Context(), flagID, req.Enabled); err != nil {
		if errors.Is(err, service.ErrFlagNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "feature flag not found"})
			return
		}
		serverError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Global Config ───────────────────────────────────────────────────────

// GetConfig handles GET /v1/platform/config/{key}
func (h *Platform) GetConfig(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	config, err := h.svc.GetConfig(r.Context(), key)
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "config key not found"})
			return
		}
		serverError(w, r, err)
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.svc.SetConfig(r.Context(), key, req.Value, adminID); err != nil {
		serverError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Middleware & Helpers ─────────────────────────────────────────────────

// requirePlatformAdmin verifies the JWT and checks that the user has admin role.
// The JWT is passed via Authorization: Bearer <token>.
func (h *Platform) requirePlatformAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization"})
			return
		}

		claims, err := sdajwt.Verify(h.publicKey, strings.TrimPrefix(auth, "Bearer "))
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}

		if claims.Role != "admin" || (h.platformSlug != "" && claims.Slug != h.platformSlug) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "platform admin access required"})
			return
		}

		// Propagate identity via headers for downstream use
		r.Header.Set("X-User-ID", claims.UserID)
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func serverError(w http.ResponseWriter, r *http.Request, err error) {
	reqID := middleware.GetReqID(r.Context())
	slog.Error("internal error", "error", err, "request_id", reqID)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}
