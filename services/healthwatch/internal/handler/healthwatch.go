// Package handler implements HTTP handlers for the healthwatch service.
// All endpoints require platform admin role verified via JWT ([DS4]).
package handler

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Camionerou/rag-saldivia/pkg/httperr"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/services/healthwatch/internal/service"
)

// HealthwatchService defines the operations the handler needs.
type HealthwatchService interface {
	Summary(ctx context.Context) (*service.HealthSummary, error)
	ServiceStatuses(ctx context.Context) ([]service.ServiceStatus, error)
	ActiveAlerts(ctx context.Context) ([]service.Alert, error)
	TriggerCheck(ctx context.Context) (*service.HealthSummary, error)
	ListTriageRecords(ctx context.Context, limit int) ([]service.TriageRecord, error)
}

// Healthwatch handles HTTP requests for health monitoring.
type Healthwatch struct {
	svc          HealthwatchService
	publicKey    ed25519.PublicKey
	platformSlug string
	blacklist    *security.TokenBlacklist
}

// New creates healthwatch HTTP handlers.
func New(svc HealthwatchService, publicKey ed25519.PublicKey, platformSlug string, blacklist *security.TokenBlacklist) *Healthwatch {
	return &Healthwatch{svc: svc, publicKey: publicKey, platformSlug: platformSlug, blacklist: blacklist}
}

// Routes returns a chi router with all healthwatch routes.
func (h *Healthwatch) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(h.requirePlatformAdmin)

	r.Get("/summary", h.Summary)
	r.Get("/services", h.Services)
	r.Get("/alerts", h.Alerts)
	r.Post("/check", h.Check)
	r.Get("/triage", h.Triage)

	return r
}

// Summary handles GET /v1/healthwatch/summary
func (h *Healthwatch) Summary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.svc.Summary(r.Context())
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// Services handles GET /v1/healthwatch/services
func (h *Healthwatch) Services(w http.ResponseWriter, r *http.Request) {
	statuses, err := h.svc.ServiceStatuses(r.Context())
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, statuses)
}

// Alerts handles GET /v1/healthwatch/alerts
func (h *Healthwatch) Alerts(w http.ResponseWriter, r *http.Request) {
	alerts, err := h.svc.ActiveAlerts(r.Context())
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, alerts)
}

// Check handles POST /v1/healthwatch/check (manual trigger, rate limited)
func (h *Healthwatch) Check(w http.ResponseWriter, r *http.Request) {
	summary, err := h.svc.TriggerCheck(r.Context())
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// Triage handles GET /v1/healthwatch/triage
func (h *Healthwatch) Triage(w http.ResponseWriter, r *http.Request) {
	records, err := h.svc.ListTriageRecords(r.Context(), 50)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	writeJSON(w, http.StatusOK, records)
}

// requirePlatformAdmin verifies JWT, checks blacklist, strips spoofed headers,
// and ensures the user has admin role on the platform tenant ([DS4]).
func (h *Healthwatch) requirePlatformAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strip spoofed identity headers
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

		// Propagate identity via headers
		r.Header.Set("X-User-ID", claims.UserID)
		r.Header.Set("X-User-Email", claims.Email)
		r.Header.Set("X-User-Role", claims.Role)
		r.Header.Set("X-Tenant-ID", claims.TenantID)
		r.Header.Set("X-Tenant-Slug", claims.Slug)
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
