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

type Invoicing struct{ svc *service.Invoicing }

func NewInvoicing(svc *service.Invoicing) *Invoicing { return &Invoicing{svc: svc} }

func (h *Invoicing) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.invoicing.read"))
		r.Get("/invoices", h.ListInvoices)
		r.Get("/invoices/{id}", h.GetInvoice)
		r.Get("/tax-book", h.GetTaxBook)
		r.Get("/withholdings", h.ListWithholdings)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.invoicing.write"))
		r.Post("/invoices", h.CreateInvoice)
		r.Post("/withholdings", h.CreateWithholding)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.invoicing.post"))
		r.Post("/invoices/{id}/post", h.PostInvoice)
	})

	// Void (cascade)
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.invoicing.read"))
		r.Get("/invoices/{id}/void-preview", h.VoidPreview)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.invoicing.void"))
		r.Post("/invoices/{id}/void", h.VoidInvoice)
	})

	return r
}

func (h *Invoicing) ListInvoices(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	q := r.URL.Query()
	invoices, err := h.svc.ListInvoices(r.Context(), slug,
		q.Get("type"), q.Get("direction"), q.Get("status"),
		pgDate(q.Get("date_from")), pgDate(q.Get("date_to")),
		p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list invoices failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"invoices": invoices})
}

func (h *Invoicing) GetInvoice(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	detail, err := h.svc.GetInvoice(r.Context(), id, slug)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

func (h *Invoicing) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var body struct {
		Number      string  `json:"number"`
		Date        string  `json:"date"`
		DueDate     *string `json:"due_date,omitempty"`
		InvoiceType string  `json:"invoice_type"`
		Direction   string  `json:"direction"`
		EntityID    string  `json:"entity_id"`
		CurrencyID  *string `json:"currency_id,omitempty"`
		OrderID     *string `json:"order_id,omitempty"`
		Lines       []struct {
			ArticleID   *string `json:"article_id,omitempty"`
			Description string  `json:"description"`
			Quantity    string  `json:"quantity"`
			UnitPrice   string  `json:"unit_price"`
			TaxRate     string  `json:"tax_rate"`
		} `json:"lines"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	entityID, err := parseUUID(body.EntityID)
	if err != nil {
		http.Error(w, `{"error":"invalid entity_id"}`, http.StatusBadRequest)
		return
	}
	var lines []service.CreateInvoiceLineRequest
	for _, l := range body.Lines {
		lines = append(lines, service.CreateInvoiceLineRequest{
			ArticleID: optUUID(l.ArticleID), Description: l.Description,
			Quantity: l.Quantity, UnitPrice: l.UnitPrice, TaxRate: l.TaxRate,
		})
	}
	var dueDate string
	if body.DueDate != nil {
		dueDate = *body.DueDate
	}
	detail, err := h.svc.CreateInvoice(r.Context(), service.CreateInvoiceRequest{
		TenantID: slug, Number: body.Number, Date: pgDate(body.Date),
		DueDate: pgDate(dueDate), InvoiceType: body.InvoiceType,
		Direction: body.Direction, EntityID: entityID,
		CurrencyID: optUUID(body.CurrencyID), OrderID: optUUID(body.OrderID),
		UserID: r.Header.Get("X-User-ID"), IP: r.RemoteAddr, Lines: lines,
	})
	if err != nil {
		slog.Error("create invoice failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(detail)
}

func (h *Invoicing) PostInvoice(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	if err := h.svc.PostInvoice(r.Context(), id, slug, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		slog.Error("post invoice failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Invoicing) GetTaxBook(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	period := r.URL.Query().Get("period")
	if period == "" {
		http.Error(w, `{"error":"period query param required (YYYY-MM)"}`, http.StatusBadRequest)
		return
	}
	direction := r.URL.Query().Get("direction")
	entries, err := h.svc.GetTaxBook(r.Context(), slug, period, direction)
	if err != nil {
		slog.Error("get tax book failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"entries": entries})
}

func (h *Invoicing) ListWithholdings(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	p := pagination.Parse(r)
	typeFilter := r.URL.Query().Get("type")
	withholdings, err := h.svc.ListWithholdings(r.Context(), slug, typeFilter, p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list withholdings failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"withholdings": withholdings})
}

func (h *Invoicing) CreateWithholding(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		InvoiceID      *string `json:"invoice_id,omitempty"`
		MovementID     *string `json:"movement_id,omitempty"`
		EntityID       string  `json:"entity_id"`
		Type           string  `json:"type"`
		Rate           string  `json:"rate"`
		BaseAmount     string  `json:"base_amount"`
		Amount         string  `json:"amount"`
		CertificateNum *string `json:"certificate_num,omitempty"`
		Date           string  `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	entityID, err := parseUUID(body.EntityID)
	if err != nil {
		http.Error(w, `{"error":"invalid entity_id"}`, http.StatusBadRequest)
		return
	}
	var certNum pgtype.Text
	if body.CertificateNum != nil && *body.CertificateNum != "" {
		certNum = pgtype.Text{String: *body.CertificateNum, Valid: true}
	}
	w2, err := h.svc.CreateWithholding(r.Context(), repository.CreateWithholdingParams{
		TenantID:       slug,
		InvoiceID:      optUUID(body.InvoiceID),
		MovementID:     optUUID(body.MovementID),
		EntityID:       entityID,
		Type:           body.Type,
		Rate:           pgNumericH(body.Rate),
		BaseAmount:     pgNumericH(body.BaseAmount),
		Amount:         pgNumericH(body.Amount),
		CertificateNum: certNum,
		Date:           pgDate(body.Date),
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("create withholding failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(w2)
}

// VoidPreview returns what voiding an invoice would do.
func (h *Invoicing) VoidPreview(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	preview, err := h.svc.VoidPreview(r.Context(), id, slug)
	if err != nil {
		slog.Error("void preview failed", "error", err)
		writeSafeErr(w, err, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preview)
}

// VoidInvoice performs cascade void on a posted/paid invoice.
func (h *Invoicing) VoidInvoice(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	var body struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	result, err := h.svc.VoidInvoice(r.Context(), id, slug, body.Reason,
		r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("void invoice failed", "error", err)
		writeSafeErr(w, err, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

