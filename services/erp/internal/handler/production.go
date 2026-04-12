package handler

import (
	"context"
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

// ProductionService is the interface the Production handler depends on.
type ProductionService interface {
	ListCenters(ctx context.Context, tenantID string) ([]repository.ErpProductionCenter, error)
	CreateCenter(ctx context.Context, tenantID, code, name, userID, ip string) (repository.ErpProductionCenter, error)
	ListOrders(ctx context.Context, tenantID, status string, limit, offset int) ([]repository.ListProductionOrdersRow, error)
	GetOrder(ctx context.Context, id pgtype.UUID, tenantID string) (*service.ProductionOrderDetail, error)
	CreateOrder(ctx context.Context, req service.CreateProductionOrderRequest) (repository.ErpProductionOrder, error)
	StartOrder(ctx context.Context, id pgtype.UUID, tenantID, userID, ip string) error
	UpdateStep(ctx context.Context, id pgtype.UUID, tenantID, status, notes, userID, ip string) error
	CreateInspection(ctx context.Context, p repository.CreateProductionInspectionParams, userID, ip string) (repository.ErpProductionInspection, error)
	ListUnits(ctx context.Context, tenantID, status string, limit, offset int) ([]repository.ListUnitsRow, error)
	GetUnit(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpUnit, error)
	CreateUnit(ctx context.Context, p repository.CreateUnitParams, userID, ip string) (repository.ErpUnit, error)
}

type Production struct{ svc ProductionService }

func NewProduction(svc ProductionService) *Production { return &Production{svc: svc} }

func (h *Production) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.production.read"))
		r.Get("/centers", h.ListCenters)
		r.Get("/orders", h.ListOrders)
		r.Get("/orders/{id}", h.GetOrder)
		r.Get("/units", h.ListUnits)
		r.Get("/units/{id}", h.GetUnit)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.production.write"))
		r.Post("/centers", h.CreateCenter)
		r.Post("/orders", h.CreateOrder)
		r.Post("/orders/{id}/start", h.StartOrder)
		r.Patch("/steps/{id}", h.UpdateStep)
		r.Post("/inspections", h.CreateInspection)
		r.Post("/units", h.CreateUnit)
	})
	return r
}

func (h *Production) ListCenters(w http.ResponseWriter, r *http.Request) {
	centers, err := h.svc.ListCenters(r.Context(), tenantSlug(r))
	if err != nil {
		slog.Error("list centers failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"centers": centers})
}

func (h *Production) CreateCenter(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct{ Code, Name string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	c, err := h.svc.CreateCenter(r.Context(), slug, body.Code, body.Name, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create center failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

func (h *Production) ListOrders(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	orders, err := h.svc.ListOrders(r.Context(), slug, r.URL.Query().Get("status"), p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list production orders failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"orders": orders})
}

func (h *Production) GetOrder(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	detail, err := h.svc.GetOrder(r.Context(), id, slug)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

func (h *Production) CreateOrder(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var body struct {
		Number    string  `json:"number"`
		Date      string  `json:"date"`
		ProductID string  `json:"product_id"`
		CenterID  *string `json:"center_id,omitempty"`
		Quantity  string  `json:"quantity"`
		Priority  int32   `json:"priority"`
		OrderID   *string `json:"order_id,omitempty"`
		StartDate *string `json:"start_date,omitempty"`
		EndDate   *string `json:"end_date,omitempty"`
		Notes     string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	prodID, err := parseUUID(body.ProductID)
	if err != nil {
		http.Error(w, `{"error":"invalid product_id"}`, http.StatusBadRequest)
		return
	}
	var sd, ed string
	if body.StartDate != nil { sd = *body.StartDate }
	if body.EndDate != nil { ed = *body.EndDate }
	order, err := h.svc.CreateOrder(r.Context(), service.CreateProductionOrderRequest{
		TenantID: slug, Number: body.Number, Date: pgDate(body.Date),
		ProductID: prodID, CenterID: optUUID(body.CenterID),
		Quantity: body.Quantity, Priority: body.Priority,
		OrderID: optUUID(body.OrderID), StartDate: pgDate(sd), EndDate: pgDate(ed),
		UserID: r.Header.Get("X-User-ID"), Notes: body.Notes, IP: r.RemoteAddr,
	})
	if err != nil {
		slog.Error("create production order failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func (h *Production) StartOrder(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	if err := h.svc.StartOrder(r.Context(), id, slug, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Production) UpdateStep(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var body struct {
		Status string `json:"status"`
		Notes  string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateStep(r.Context(), id, slug, body.Status, body.Notes, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Production) CreateInspection(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		OrderID     string  `json:"order_id"`
		StepID      *string `json:"step_id,omitempty"`
		InspectorID *string `json:"inspector_id,omitempty"`
		Result      string  `json:"result"`
		Observations string `json:"observations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	orderID, err := parseUUID(body.OrderID)
	if err != nil {
		http.Error(w, `{"error":"invalid order_id"}`, http.StatusBadRequest)
		return
	}
	insp, err := h.svc.CreateInspection(r.Context(), repository.CreateProductionInspectionParams{
		TenantID: slug, OrderID: orderID, StepID: optUUID(body.StepID),
		InspectorID: optUUID(body.InspectorID), Result: body.Result,
		Observations: body.Observations,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create inspection failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(insp)
}

func (h *Production) ListUnits(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	units, err := h.svc.ListUnits(r.Context(), slug, r.URL.Query().Get("status"), p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list units failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"units": units})
}

func (h *Production) GetUnit(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	unit, err := h.svc.GetUnit(r.Context(), id, slug)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(unit)
}

func (h *Production) CreateUnit(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var body struct {
		ChassisNumber    string  `json:"chassis_number"`
		InternalNumber   *string `json:"internal_number,omitempty"`
		Model            *string `json:"model,omitempty"`
		CustomerID       *string `json:"customer_id,omitempty"`
		OrderID          *string `json:"order_id,omitempty"`
		ProductionOrderID *string `json:"production_order_id,omitempty"`
		Patent           *string `json:"patent,omitempty"`
		EngineBrand      *string `json:"engine_brand,omitempty"`
		BodyStyle        *string `json:"body_style,omitempty"`
		SeatCount        *int32  `json:"seat_count,omitempty"`
		Year             *int32  `json:"year,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	var sc, yr pgtype.Int4
	if body.SeatCount != nil { sc = pgtype.Int4{Int32: *body.SeatCount, Valid: true} }
	if body.Year != nil { yr = pgtype.Int4{Int32: *body.Year, Valid: true} }

	unit, err := h.svc.CreateUnit(r.Context(), repository.CreateUnitParams{
		TenantID:          slug,
		ChassisNumber:     body.ChassisNumber,
		InternalNumber:    pgTextOpt(body.InternalNumber),
		Model:             pgTextOpt(body.Model),
		CustomerID:        optUUID(body.CustomerID),
		OrderID:           optUUID(body.OrderID),
		ProductionOrderID: optUUID(body.ProductionOrderID),
		Patent:            pgTextOpt(body.Patent),
		EngineBrand:       pgTextOpt(body.EngineBrand),
		BodyStyle:         pgTextOpt(body.BodyStyle),
		SeatCount:         sc,
		Year:              yr,
		Metadata:          []byte(`{}`),
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create unit failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(unit)
}

