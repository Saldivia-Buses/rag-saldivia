// Package handler provides HTTP handlers for the Traces Service.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Camionerou/rag-saldivia/services/traces/internal/service"
)

// Handler wraps the Traces service for HTTP.
type Handler struct {
	svc *service.Traces
}

// New creates a traces Handler.
func New(svc *service.Traces) *Handler {
	return &Handler{svc: svc}
}

// Routes returns the traces router.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListTraces)
	r.Get("/{traceID}", h.GetTraceDetail)
	r.Get("/costs/{tenantID}", h.GetTenantCost)
	return r
}

// ListTraces handles GET /v1/traces?tenant_id=X&limit=N&offset=N
func (h *Handler) ListTraces(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, `{"error":"tenant_id is required"}`, http.StatusBadRequest)
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
func (h *Handler) GetTraceDetail(w http.ResponseWriter, r *http.Request) {
	traceID := chi.URLParam(r, "traceID")

	trace, events, err := h.svc.GetTraceDetail(r.Context(), traceID)
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

// GetTenantCost handles GET /v1/traces/costs/{tenantID}?from=2026-04-01&to=2026-05-01
func (h *Handler) GetTenantCost(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")

	from, err := time.Parse("2006-01-02", r.URL.Query().Get("from"))
	if err != nil {
		from = time.Now().AddDate(0, -1, 0) // default: last month
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
