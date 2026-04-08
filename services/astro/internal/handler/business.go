package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/business"
)

// --- Business Intelligence Endpoints ---
// All behind astro.business permission.

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if h.q == nil || h.biz == nil {
		jsonError(w, "business module not configured", http.StatusServiceUnavailable)
		return
	}

	year, month := parseYearMonth(r)
	companyID := r.URL.Query().Get("company_id")
	if companyID == "" {
		jsonError(w, "company_id is required", http.StatusBadRequest)
		return
	}

	// Resolve company contact
	contact, code, err := h.resolveContact(r, companyID)
	if err != nil {
		jsonError(w, err.Error(), code)
		return
	}
	if contact.Kind != "empresa" {
		jsonError(w, "contact must be kind=empresa", http.StatusBadRequest)
		return
	}

	chart, _, err := contactToChart(contact)
	if err != nil {
		serverError(w, r, "company chart failed", err)
		return
	}

	// For now, no counterparty charts (would need additional query params)
	dashboard := h.biz.BuildDashboard(chart, contact.Name, nil, year, month)
	jsonOK(w, dashboard)
}

func (h *Handler) CashFlow(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}

	year, _ := parseYearMonth(r)
	companyID := r.URL.Query().Get("company_id")
	if companyID == "" {
		jsonError(w, "company_id is required", http.StatusBadRequest)
		return
	}

	contact, code, err := h.resolveContact(r, companyID)
	if err != nil {
		jsonError(w, err.Error(), code)
		return
	}
	chart, _, err := contactToChart(contact)
	if err != nil {
		serverError(w, r, "chart failed", err)
		return
	}

	jsonOK(w, business.CalcCashFlow(chart, year))
}

func (h *Handler) RiskHeatmap(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}

	year, _ := parseYearMonth(r)
	companyID := r.URL.Query().Get("company_id")
	if companyID == "" {
		jsonError(w, "company_id is required", http.StatusBadRequest)
		return
	}

	contact, code, err := h.resolveContact(r, companyID)
	if err != nil {
		jsonError(w, err.Error(), code)
		return
	}
	chart, _, err := contactToChart(contact)
	if err != nil {
		serverError(w, r, "chart failed", err)
		return
	}

	jsonOK(w, business.CalcRiskHeatmap(chart, year))
}

func (h *Handler) MercuryRx(w http.ResponseWriter, r *http.Request) {
	year, _ := parseYearMonth(r)
	jsonOK(w, business.CalcMercuryRx(year))
}

func (h *Handler) QuarterlyForecast(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}

	year, month := parseYearMonth(r)
	quarter := (month-1)/3 + 1
	companyID := r.URL.Query().Get("company_id")
	if companyID == "" {
		jsonError(w, "company_id is required", http.StatusBadRequest)
		return
	}

	contact, code, err := h.resolveContact(r, companyID)
	if err != nil {
		jsonError(w, err.Error(), code)
		return
	}
	chart, _, err := contactToChart(contact)
	if err != nil {
		serverError(w, r, "chart failed", err)
		return
	}

	jsonOK(w, business.CalcQuarterlyForecast(chart, year, quarter))
}

func (h *Handler) TeamCompatibility(w http.ResponseWriter, r *http.Request) {
	jsonError(w, "team compatibility requires multiple contacts — use synastry endpoint", http.StatusNotImplemented)
}

func (h *Handler) HiringCalendar(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}

	year, month := parseYearMonth(r)
	companyID := r.URL.Query().Get("company_id")
	if companyID == "" {
		jsonError(w, "company_id is required", http.StatusBadRequest)
		return
	}

	contact, code, err := h.resolveContact(r, companyID)
	if err != nil {
		jsonError(w, err.Error(), code)
		return
	}
	chart, _, err := contactToChart(contact)
	if err != nil {
		serverError(w, r, "chart failed", err)
		return
	}

	jsonOK(w, business.CalcHiringCalendar(chart, year, month))
}

// parseYearMonth extracts year and month from query params, defaulting to now.
func parseYearMonth(r *http.Request) (int, int) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	if v := r.URL.Query().Get("year"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1900 && n <= 2200 {
			year = n
		}
	}
	if v := r.URL.Query().Get("month"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 12 {
			month = n
		}
	}
	return year, month
}
