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

// SalesService is the interface the Sales handler depends on.
type SalesService interface {
	ListQuotations(ctx context.Context, tenantID, status string, limit, offset int) ([]repository.ListQuotationsRow, error)
	GetQuotation(ctx context.Context, id pgtype.UUID, tenantID string) (*service.QuotationDetail, error)
	CreateQuotation(ctx context.Context, req service.CreateQuotationRequest) (*service.QuotationDetail, error)
	ListOrders(ctx context.Context, tenantID, status, orderType string, limit, offset int) ([]repository.ListOrdersRow, error)
	GetOrder(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpOrder, error)
	CreateOrder(ctx context.Context, tenantID, number string, date pgtype.Date, orderType string, customerID, quotationID pgtype.UUID, total pgtype.Numeric, notes, userID, ip string) (repository.ErpOrder, error)
	UpdateOrderStatus(ctx context.Context, id pgtype.UUID, tenantID, status, userID, ip string) error
	ListPriceLists(ctx context.Context, tenantID string) ([]repository.ErpPriceList, error)
	GetPriceList(ctx context.Context, id pgtype.UUID, tenantID string) (*service.PriceListDetail, error)
}

type Sales struct{ svc SalesService }

func NewSales(svc SalesService) *Sales { return &Sales{svc: svc} }

func (h *Sales) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.sales.read"))
		r.Get("/quotations", h.ListQuotations)
		r.Get("/quotations/{id}", h.GetQuotation)
		r.Get("/orders", h.ListOrders)
		r.Get("/orders/{id}", h.GetOrder)
		r.Get("/price-lists", h.ListPriceLists)
		r.Get("/price-lists/{id}", h.GetPriceList)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.sales.write"))
		r.Post("/quotations", h.CreateQuotation)
		r.Post("/orders", h.CreateOrder)
		r.Patch("/orders/{id}/status", h.UpdateOrderStatus)
	})
	return r
}

func (h *Sales) ListQuotations(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	status := r.URL.Query().Get("status")
	quotations, err := h.svc.ListQuotations(r.Context(), slug, status, p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"quotations": quotations})
}

func (h *Sales) GetQuotation(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	detail, err := h.svc.GetQuotation(r.Context(), id, slug)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("quotation"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(detail)
}

func (h *Sales) CreateQuotation(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var body struct {
		Number     string  `json:"number"`
		Date       string  `json:"date"`
		CustomerID string  `json:"customer_id"`
		CurrencyID *string `json:"currency_id,omitempty"`
		ValidUntil *string `json:"valid_until,omitempty"`
		Notes      string  `json:"notes"`
		Lines      []struct {
			ArticleID   *string `json:"article_id,omitempty"`
			Description string  `json:"description"`
			Quantity    string  `json:"quantity"`
			UnitPrice   string  `json:"unit_price"`
		} `json:"lines"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}
	custID, err := parseUUID(body.CustomerID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID("customer_id"))
		return
	}
	var lines []service.CreateQuotationLineRequest
	for _, l := range body.Lines {
		lines = append(lines, service.CreateQuotationLineRequest{
			ArticleID: optUUID(l.ArticleID), Description: l.Description,
			Quantity: l.Quantity, UnitPrice: l.UnitPrice,
		})
	}
	var validUntil string
	if body.ValidUntil != nil {
		validUntil = *body.ValidUntil
	}
	detail, err := h.svc.CreateQuotation(r.Context(), service.CreateQuotationRequest{
		TenantID: slug, Number: body.Number, Date: pgDate(body.Date),
		CustomerID: custID, CurrencyID: optUUID(body.CurrencyID),
		ValidUntil: pgDate(validUntil), Notes: body.Notes,
		UserID: r.Header.Get("X-User-ID"), IP: r.RemoteAddr, Lines: lines,
	})
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(detail)
}

func (h *Sales) ListOrders(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	orders, err := h.svc.ListOrders(r.Context(), slug, r.URL.Query().Get("status"), r.URL.Query().Get("type"), p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"orders": orders})
}

func (h *Sales) GetOrder(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	order, err := h.svc.GetOrder(r.Context(), id, slug)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("order"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(order)
}

func (h *Sales) CreateOrder(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		Number      string  `json:"number"`
		Date        string  `json:"date"`
		OrderType   string  `json:"order_type"`
		CustomerID  *string `json:"customer_id,omitempty"`
		QuotationID *string `json:"quotation_id,omitempty"`
		Total       string  `json:"total"`
		Notes       string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}
	order, err := h.svc.CreateOrder(r.Context(), slug, body.Number, pgDate(body.Date),
		body.OrderType, optUUID(body.CustomerID), optUUID(body.QuotationID),
		pgNumericH(body.Total), body.Notes, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(order)
}

func (h *Sales) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	var body struct{ Status string `json:"status"` }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid request body"))
		return
	}
	if err := h.svc.UpdateOrderStatus(r.Context(), id, slug, body.Status, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("order"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Sales) ListPriceLists(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	lists, err := h.svc.ListPriceLists(r.Context(), slug)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"price_lists": lists})
}

func (h *Sales) GetPriceList(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	detail, err := h.svc.GetPriceList(r.Context(), id, slug)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("price_list"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(detail)
}
