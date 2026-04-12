// Package handler provides HTTP handlers for the Traces Service.
package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Camionerou/rag-saldivia/pkg/tenant"
	"github.com/Camionerou/rag-saldivia/services/traces/internal/service"
)

// TracesService is the narrow interface the handler needs.
type TracesService interface {
	ListTraces(ctx context.Context, tenantID string, limit, offset int) ([]service.Trace, error)
	GetTraceDetail(ctx context.Context, traceID, tenantID string) (*service.Trace, []service.TraceEvent, error)
	GetTenantCost(ctx context.Context, tenantID string, from, to time.Time) (*service.CostSummary, error)
}

// Handler wraps the Traces service for HTTP.
type Handler struct {
	svc TracesService
}

// New creates a traces Handler.
func New(svc *service.Traces) *Handler {
	return &Handler{svc: svc}
}

// Routes returns the traces router.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListTraces)
	r.Get("/costs", h.GetTenantCost)
	r.Get("/{traceID}", h.GetTraceDetail)
	return r
}

// ListTraces handles GET /v1/traces?limit=N&offset=N
// B2: tenant_id comes from JWT context, not query string
func (h *Handler) ListTraces(w http.ResponseWriter, r *http.Request) {
	ti, err := tenant.FromContext(r.Context())
	tenantID := ti.ID
	if err != nil || tenantID == "" {
		http.Error(w, `{"error":"tenant context missing"}`, http.StatusUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	traces, err := h.svc.ListTraces(r.Context(), tenantID, limit, offset)
	if err != nil {
		slog.Error("list traces failed", "error", err)
		http.Error(w, `{"error":"failed to list traces"}`, http.StatusInternalServerError)
		return
	}
	if traces == nil {
		traces = []service.Trace{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(traces)
}

// GetTraceDetail handles GET /v1/traces/{traceID}
// B3: enforces tenant isolation via tenant_id in query
func (h *Handler) GetTraceDetail(w http.ResponseWriter, r *http.Request) {
	ti, err := tenant.FromContext(r.Context())
	tenantID := ti.ID
	if err != nil || tenantID == "" {
		http.Error(w, `{"error":"tenant context missing"}`, http.StatusUnauthorized)
		return
	}

	traceID := chi.URLParam(r, "traceID")

	trace, events, err := h.svc.GetTraceDetail(r.Context(), traceID, tenantID)
	if err != nil {
		slog.Error("get trace failed", "error", err, "trace_id", traceID)
		http.Error(w, `{"error":"trace not found"}`, http.StatusNotFound)
		return
	}

	resp := map[string]any{
		"trace":  trace,
		"events": events,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetTenantCost handles GET /v1/traces/costs?from=2026-04-01&to=2026-05-01
// B2: tenant_id from JWT, not URL param
func (h *Handler) GetTenantCost(w http.ResponseWriter, r *http.Request) {
	ti, err := tenant.FromContext(r.Context())
	tenantID := ti.ID
	if err != nil || tenantID == "" {
		http.Error(w, `{"error":"tenant context missing"}`, http.StatusUnauthorized)
		return
	}

	from, err := time.Parse("2006-01-02", r.URL.Query().Get("from"))
	if err != nil {
		from = time.Now().AddDate(0, -1, 0)
	}
	to, err := time.Parse("2006-01-02", r.URL.Query().Get("to"))
	if err != nil {
		to = time.Now()
	}

	cost, err := h.svc.GetTenantCost(r.Context(), tenantID, from, to)
	if err != nil {
		slog.Error("get cost failed", "error", err, "tenant_id", tenantID)
		http.Error(w, `{"error":"failed to get costs"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cost)
}
