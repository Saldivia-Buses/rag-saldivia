package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/service"
)

type Maintenance struct{ svc *service.Maintenance }

func NewMaintenance(svc *service.Maintenance) *Maintenance { return &Maintenance{svc: svc} }

func (h *Maintenance) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.maintenance.read"))
		r.Get("/assets", h.ListAssets)
		r.Get("/assets/{id}/plans", h.ListPlans)
		r.Get("/work-orders", h.ListWorkOrders)
		r.Get("/work-orders/{id}", h.GetWorkOrder)
		r.Get("/fuel-logs", h.ListFuelLogs)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.maintenance.write"))
		r.Post("/assets", h.CreateAsset)
		r.Post("/assets/{id}/plans", h.CreatePlan)
		r.Post("/work-orders", h.CreateWorkOrder)
		r.Patch("/work-orders/{id}/status", h.UpdateWorkOrderStatus)
		r.Post("/fuel-logs", h.CreateFuelLog)
	})
	return r
}

func (h *Maintenance) ListAssets(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	q := r.URL.Query()
	assets, err := h.svc.ListAssets(r.Context(), slug, q.Get("type"), q.Get("active") != "false")
	if err != nil { slog.Error("list assets failed", "error", err); http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return }
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"assets": assets})
}

func (h *Maintenance) CreateAsset(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct{ Code, Name, AssetType, Location string; UnitID *string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest); return }
	a, err := h.svc.CreateAsset(r.Context(), repository.CreateMaintenanceAssetParams{
		TenantID: slug, Code: body.Code, Name: body.Name, AssetType: body.AssetType,
		UnitID: optUUID(body.UnitID), Location: body.Location, Metadata: []byte(`{}`),
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil { slog.Error("create asset failed", "error", err); http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return }
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(http.StatusCreated); json.NewEncoder(w).Encode(a)
}

func (h *Maintenance) ListPlans(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil { http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest); return }
	plans, err := h.svc.ListPlans(r.Context(), tenantSlug(r), id)
	if err != nil { slog.Error("list plans failed", "error", err); http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return }
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(map[string]any{"plans": plans})
}

func (h *Maintenance) CreatePlan(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r); r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	assetID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil { http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest); return }
	var body struct{ Name string; FreqDays, FreqKm, FreqHours *int32; NextDue *string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest); return }
	var nd string; if body.NextDue != nil { nd = *body.NextDue }
	pl, err := h.svc.CreatePlan(r.Context(), repository.CreateMaintenancePlanParams{
		TenantID: slug, AssetID: assetID, Name: body.Name,
		FrequencyDays: optInt4(body.FreqDays), FrequencyKm: optInt4(body.FreqKm),
		FrequencyHours: optInt4(body.FreqHours), NextDue: pgDate(nd),
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil { slog.Error("create plan failed", "error", err); http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return }
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(http.StatusCreated); json.NewEncoder(w).Encode(pl)
}

func (h *Maintenance) ListWorkOrders(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r); p := pagination.Parse(r)
	wos, err := h.svc.ListWorkOrders(r.Context(), slug, r.URL.Query().Get("status"), p.Limit(), p.Offset())
	if err != nil { slog.Error("list work orders failed", "error", err); http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return }
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(map[string]any{"work_orders": wos})
}

func (h *Maintenance) GetWorkOrder(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil { http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest); return }
	detail, err := h.svc.GetWorkOrder(r.Context(), id, tenantSlug(r))
	if err != nil { http.Error(w, `{"error":"not found"}`, http.StatusNotFound); return }
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(detail)
}

func (h *Maintenance) CreateWorkOrder(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r); r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		Number, WorkType, Description, Notes string
		AssetID string; PlanID, AssignedTo *string
		Date string; Priority string
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest); return }
	assetID, err := parseUUID(body.AssetID)
	if err != nil { http.Error(w, `{"error":"invalid asset_id"}`, http.StatusBadRequest); return }
	if body.Priority == "" { body.Priority = "normal" }
	wo, err := h.svc.CreateWorkOrder(r.Context(), repository.CreateWorkOrderParams{
		TenantID: slug, Number: body.Number, AssetID: assetID, PlanID: optUUID(body.PlanID),
		Date: pgDate(body.Date), WorkType: body.WorkType, Description: body.Description,
		AssignedTo: optUUID(body.AssignedTo), Priority: body.Priority,
		UserID: r.Header.Get("X-User-ID"), Notes: body.Notes,
	}, r.RemoteAddr)
	if err != nil { slog.Error("create work order failed", "error", err); http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return }
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(http.StatusCreated); json.NewEncoder(w).Encode(wo)
}

func (h *Maintenance) UpdateWorkOrderStatus(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r); r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil { http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest); return }
	var body struct{ Status string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest); return }
	if err := h.svc.UpdateWorkOrderStatus(r.Context(), id, slug, body.Status, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		if err.Error() == "work order not found" { http.Error(w, `{"error":"not found"}`, http.StatusNotFound) } else if err.Error() == "invalid status" { http.Error(w, `{"error":"invalid status"}`, http.StatusBadRequest) } else { slog.Error("update wo status failed", "error", err); http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError) }
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Maintenance) ListFuelLogs(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r); p := pagination.Parse(r)
	logs, err := h.svc.ListFuelLogs(r.Context(), slug, optUUID(ptrStr(r.URL.Query().Get("asset_id"))), p.Limit(), p.Offset())
	if err != nil { slog.Error("list fuel logs failed", "error", err); http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return }
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(map[string]any{"fuel_logs": logs})
}

func (h *Maintenance) CreateFuelLog(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r); r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct { AssetID, Date string; Liters, Cost string; KmReading *int32 }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest); return }
	assetID, err := parseUUID(body.AssetID)
	if err != nil { http.Error(w, `{"error":"invalid asset_id"}`, http.StatusBadRequest); return }
	fl, err := h.svc.CreateFuelLog(r.Context(), repository.CreateFuelLogParams{
		TenantID: slug, AssetID: assetID, Date: pgDate(body.Date),
		Liters: pgNumericH(body.Liters), KmReading: optInt4(body.KmReading),
		Cost: pgNumericH(body.Cost), UserID: r.Header.Get("X-User-ID"),
	}, r.RemoteAddr)
	if err != nil { slog.Error("create fuel log failed", "error", err); http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError); return }
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(http.StatusCreated); json.NewEncoder(w).Encode(fl)
}

func optInt4(v *int32) pgtype.Int4 {
	if v == nil { return pgtype.Int4{} }
	return pgtype.Int4{Int32: *v, Valid: true}
}
