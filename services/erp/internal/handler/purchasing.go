package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	erperrors "github.com/Camionerou/rag-saldivia/services/erp/internal/errors"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/service"
)

// PurchasingService is the interface the Purchasing handler depends on.
type PurchasingService interface {
	ListOrders(ctx context.Context, tenantID, status string, limit, offset int) ([]repository.ListPurchaseOrdersRow, error)
	GetOrder(ctx context.Context, id pgtype.UUID, tenantID string) (*service.OrderDetail, error)
	CreateOrder(ctx context.Context, req service.CreateOrderRequest) (*service.OrderDetail, error)
	ApproveOrder(ctx context.Context, id pgtype.UUID, tenantID, userID, ip string) error
	Receive(ctx context.Context, req service.ReceiveRequest) error
	ListReceipts(ctx context.Context, tenantID string, limit, offset int) ([]repository.ListPurchaseReceiptsRow, error)
	InspectReceipt(ctx context.Context, tenantID string, receiptID pgtype.UUID, inspections []service.InspectionInput, inspectorID, ip string) ([]repository.ErpQcInspection, error)
	ListInspections(ctx context.Context, tenantID, status string, limit, offset int) ([]repository.ListInspectionsRow, error)
	GetInspection(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpQcInspection, error)
	ListSupplierDemerits(ctx context.Context, tenantID string, supplierID pgtype.UUID) ([]repository.ErpSupplierDemerit, error)
	GetSupplierDemeritTotal(ctx context.Context, tenantID string, supplierID pgtype.UUID) (int32, error)
}

type Purchasing struct{ svc PurchasingService }

func NewPurchasing(svc PurchasingService) *Purchasing { return &Purchasing{svc: svc} }

func (h *Purchasing) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.purchasing.read"))
		r.Get("/orders", h.ListOrders)
		r.Get("/orders/{id}", h.GetOrder)
		r.Get("/receipts", h.ListReceipts)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.purchasing.write"))
		r.Post("/orders", h.CreateOrder)
		r.Post("/orders/{id}/receive", h.Receive)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.purchasing.approve"))
		r.Post("/orders/{id}/approve", h.Approve)
	})

	// QC Inspections
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.purchasing.read"))
		r.Get("/inspections", h.ListInspections)
		r.Get("/inspections/{id}", h.GetInspection)
		r.Get("/suppliers/{id}/demerits", h.ListSupplierDemerits)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.purchasing.inspect"))
		r.Post("/receipts/{id}/inspect", h.InspectReceipt)
	})

	return r
}

func (h *Purchasing) ListOrders(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	status := r.URL.Query().Get("status")
	orders, err := h.svc.ListOrders(r.Context(), slug, status, p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"orders": orders})
}

func (h *Purchasing) GetOrder(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	detail, err := h.svc.GetOrder(r.Context(), id, slug)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("order"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

func (h *Purchasing) CreateOrder(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var body struct {
		Number     string  `json:"number"`
		Date       string  `json:"date"`
		SupplierID string  `json:"supplier_id"`
		CurrencyID *string `json:"currency_id,omitempty"`
		Notes      string  `json:"notes"`
		Lines      []struct {
			ArticleID string `json:"article_id"`
			Quantity  string `json:"quantity"`
			UnitPrice string `json:"unit_price"`
		} `json:"lines"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}
	supplierID, err := parseUUID(body.SupplierID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID("supplier_id"))
		return
	}
	var lines []service.CreateOrderLineRequest
	for _, l := range body.Lines {
		artID, err := parseUUID(l.ArticleID)
		if err != nil {
			erperrors.WriteError(w, r, erperrors.InvalidID("article_id in line"))
			return
		}
		lines = append(lines, service.CreateOrderLineRequest{
			ArticleID: artID, Quantity: l.Quantity, UnitPrice: l.UnitPrice,
		})
	}
	detail, err := h.svc.CreateOrder(r.Context(), service.CreateOrderRequest{
		TenantID:   slug,
		Number:     body.Number,
		Date:       pgDate(body.Date),
		SupplierID: supplierID,
		CurrencyID: optUUID(body.CurrencyID),
		Notes:      body.Notes,
		UserID:     r.Header.Get("X-User-ID"),
		IP:         r.RemoteAddr,
		Lines:      lines,
	})
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(detail)
}

func (h *Purchasing) Approve(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	if err := h.svc.ApproveOrder(r.Context(), id, slug, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("order"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Purchasing) Receive(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	orderID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	var body struct {
		Date   string `json:"date"`
		Number string `json:"number"`
		Notes  string `json:"notes"`
		Lines  []struct {
			OrderLineID string `json:"order_line_id"`
			ArticleID   string `json:"article_id"`
			Quantity    string `json:"quantity"`
		} `json:"lines"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}
	var lines []service.ReceiveLineRequest
	for _, l := range body.Lines {
		olID, err := parseUUID(l.OrderLineID)
		if err != nil {
			erperrors.WriteError(w, r, erperrors.InvalidID("order_line_id in line"))
			return
		}
		aID, err := parseUUID(l.ArticleID)
		if err != nil {
			erperrors.WriteError(w, r, erperrors.InvalidID("article_id in line"))
			return
		}
		lines = append(lines, service.ReceiveLineRequest{
			OrderLineID: olID, ArticleID: aID, Quantity: l.Quantity,
		})
	}
	if err := h.svc.Receive(r.Context(), service.ReceiveRequest{
		TenantID: slug, OrderID: orderID, Date: pgDate(body.Date),
		Number: body.Number, UserID: r.Header.Get("X-User-ID"),
		Notes: body.Notes, IP: r.RemoteAddr, Lines: lines,
	}); err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Purchasing) InspectReceipt(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	receiptID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var body struct {
		Inspections []service.InspectionInput `json:"inspections"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Inspections) == 0 {
		erperrors.WriteError(w, r, erperrors.InvalidInput("inspections array required"))
		return
	}
	results, err := h.svc.InspectReceipt(r.Context(), slug, receiptID, body.Inspections,
		r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"inspections": results})
}

func (h *Purchasing) ListInspections(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	status := r.URL.Query().Get("status")
	inspections, err := h.svc.ListInspections(r.Context(), slug, status, p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"inspections": inspections})
}

func (h *Purchasing) GetInspection(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	insp, err := h.svc.GetInspection(r.Context(), id, slug)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("inspection"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(insp)
}

func (h *Purchasing) ListSupplierDemerits(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	supplierID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	demerits, err := h.svc.ListSupplierDemerits(r.Context(), slug, supplierID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	total, _ := h.svc.GetSupplierDemeritTotal(r.Context(), slug, supplierID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"demerits": demerits, "total_points": total})
}

func (h *Purchasing) ListReceipts(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	receipts, err := h.svc.ListReceipts(r.Context(), slug, p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"receipts": receipts})
}
