package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/export"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	erperrors "github.com/Camionerou/rag-saldivia/services/erp/internal/errors"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

// AnalyticsService is the interface the Analytics handler depends on.
type AnalyticsService interface {
	Repo() *repository.Queries
}

type Analytics struct{ svc AnalyticsService }

func NewAnalytics(svc AnalyticsService) *Analytics { return &Analytics{svc: svc} }

func (h *Analytics) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	// Accounting reports
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.accounting.read"))
		r.Get("/accounting/balance-evolution", h.wrap("balance_evolution", balanceEvolutionCols, h.fetchBalanceEvolution))
		r.Get("/accounting/income-expense", h.wrap("income_expense", incomeExpenseCols, h.fetchIncomeExpense))
		r.Get("/accounting/cost-center-breakdown", h.wrap("cost_center_breakdown", costCenterCols, h.fetchCostCenterBreakdown))
	})

	// Treasury reports
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.treasury.read"))
		r.Get("/treasury/cash-flow", h.wrap("cash_flow", cashFlowCols, h.fetchCashFlow))
		r.Get("/treasury/balance-evolution", h.wrap("treasury_balance", treasuryBalanceCols, h.fetchTreasuryBalance))
		r.Get("/treasury/payment-methods", h.wrap("payment_methods", paymentMethodCols, h.fetchPaymentMethods))
	})

	// Invoicing reports
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.invoicing.read"))
		r.Get("/invoicing/monthly-totals", h.wrap("monthly_totals", monthlyTotalsCols, h.fetchMonthlyTotals))
		r.Get("/invoicing/tax-summary", h.wrap("tax_summary", taxSummaryCols, h.fetchTaxSummary))
		r.Get("/invoicing/top-customers", h.wrap("top_customers", topEntityCols, h.fetchTopCustomers))
		r.Get("/invoicing/top-suppliers", h.wrap("top_suppliers", topEntityCols, h.fetchTopSuppliers))
	})

	// Current accounts reports
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.accounts.read"))
		r.Get("/accounts/aging", h.wrap("aging", agingCols, h.fetchAging))
		r.Get("/accounts/collection-rate", h.wrap("collection_rate", collectionCols, h.fetchCollectionRate))
	})

	// Stock reports
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.stock.read"))
		r.Get("/stock/valuation", h.wrap("stock_valuation", stockValuationCols, h.fetchStockValuation))
		r.Get("/stock/rotation", h.wrap("stock_rotation", stockRotationCols, h.fetchStockRotation))
		r.Get("/stock/below-minimum", h.wrap("below_minimum", belowMinCols, h.fetchBelowMinimum))
		r.Get("/stock/movement-summary", h.wrap("movement_summary", movementSummaryCols, h.fetchMovementSummary))
	})

	// Production reports
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.production.read"))
		r.Get("/production/orders-by-status", h.wrap("prod_by_status", prodStatusCols, h.fetchProdByStatus))
		r.Get("/production/output-by-month", h.wrap("prod_by_month", prodMonthCols, h.fetchProdByMonth))
		r.Get("/production/efficiency", h.wrap("prod_efficiency", prodEfficiencyCols, h.fetchProdEfficiency))
	})

	// Purchasing reports
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.purchasing.read"))
		r.Get("/purchasing/orders-by-month", h.wrap("purchasing_monthly", purchasingMonthlyCols, h.fetchPurchasingMonthly))
		r.Get("/purchasing/supplier-ranking", h.wrap("supplier_ranking", supplierRankingCols, h.fetchSupplierRanking))
	})

	// Sales reports
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.sales.read"))
		r.Get("/sales/orders-by-month", h.wrap("sales_monthly", salesMonthlyCols, h.fetchSalesMonthly))
		r.Get("/sales/quotation-conversion", h.wrap("quotation_conversion", conversionCols, h.fetchQuotationConversion))
		r.Get("/sales/customer-concentration", h.wrap("customer_concentration", customerConcentrationCols, h.fetchCustomerConcentration))
	})

	// HR reports
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.hr.read"))
		r.Get("/hr/headcount", h.wrap("headcount", headcountCols, h.fetchHeadcount))
		r.Get("/hr/absences-by-month", h.wrap("absences", absenceCols, h.fetchAbsences))
		r.Get("/hr/overtime-summary", h.wrap("overtime", overtimeCols, h.fetchOvertime))
	})

	// Quality reports
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.quality.read"))
		r.Get("/quality/nonconformities-by-type", h.wrap("nc_by_type", ncTypeCols, h.fetchNCByType))
		r.Get("/quality/resolution-time", h.wrap("nc_resolution", ncResolutionCols, h.fetchNCResolution))
	})

	// Maintenance reports
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.maintenance.read"))
		r.Get("/maintenance/completion-rate", h.wrap("wo_completion", woCompletionCols, h.fetchWOCompletion))
		r.Get("/maintenance/cost-by-asset", h.wrap("cost_by_asset", costByAssetCols, h.fetchCostByAsset))
	})

	// Dashboard KPIs (requires minimal permission)
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.catalogs.read"))
		r.Get("/dashboard/kpis", h.DashboardKPIs)
	})

	return r
}

// ─── Report wrapper ────────────────────────────────────────────────────────

type fetchFn func(r *http.Request, slug string) ([]export.Row, error)

func (h *Analytics) wrap(name string, cols []export.Column, fn fetchFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := tenantSlug(r)
		rows, err := fn(r, slug)
		if err != nil {
			erperrors.WriteError(w, r, erperrors.Internal(err))
			return
		}

		format := r.URL.Query().Get("format")
		switch format {
		case "csv":
			w.Header().Set("Content-Type", "text/csv")
			w.Header().Set("Content-Disposition", "attachment; filename="+name+".csv")
			if err := export.WriteCSV(w, cols, rows); err != nil {
				slog.Error("csv export failed", "report", name, "error", err)
			}
		case "excel":
			w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
			w.Header().Set("Content-Disposition", "attachment; filename="+name+".xlsx")
			if err := export.WriteExcel(w, "Reporte", cols, rows); err != nil {
				slog.Error("excel export failed", "report", name, "error", err)
			}
		default:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"columns": cols,
				"rows":    rows,
				"meta":    map[string]any{"count": len(rows), "report": name},
			})
		}
	}
}

// ─── Date range helper ─────────────────────────────────────────────────────

func parseTsRange(r *http.Request) (pgtype.Timestamptz, pgtype.Timestamptz) {
	from, to := parseDateRange(r)
	return pgtype.Timestamptz{Time: from.Time, Valid: true}, pgtype.Timestamptz{Time: to.Time, Valid: true}
}

func parseDateRange(r *http.Request) (pgtype.Date, pgtype.Date) {
	now := time.Now()
	from := time.Date(now.Year()-1, now.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := now

	if v := r.URL.Query().Get("date_from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			from = t
		}
	}
	if v := r.URL.Query().Get("date_to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			to = t
		}
	}

	return pgtype.Date{Time: from, Valid: true}, pgtype.Date{Time: to, Valid: true}
}

func parseTopN(r *http.Request) int32 {
	if v := r.URL.Query().Get("top"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			return int32(n)
		}
	}
	return 10
}

// ─── Column definitions ────────────────────────────────────────────────────

var (
	balanceEvolutionCols = []export.Column{
		{Header: "Período", Key: "period", Format: "text"},
		{Header: "Tipo", Key: "account_type", Format: "text"},
		{Header: "Debe", Key: "total_debit", Format: "currency"},
		{Header: "Haber", Key: "total_credit", Format: "currency"},
		{Header: "Neto", Key: "net", Format: "currency"},
	}
	incomeExpenseCols = []export.Column{
		{Header: "Período", Key: "period", Format: "text"},
		{Header: "Ingresos", Key: "income", Format: "currency"},
		{Header: "Egresos", Key: "expense", Format: "currency"},
	}
	costCenterCols = []export.Column{
		{Header: "Centro de Costo", Key: "cost_center", Format: "text"},
		{Header: "Tipo", Key: "account_type", Format: "text"},
		{Header: "Total", Key: "total", Format: "currency"},
	}
	cashFlowCols = []export.Column{
		{Header: "Período", Key: "period", Format: "text"},
		{Header: "Ingresos", Key: "inflow", Format: "currency"},
		{Header: "Egresos", Key: "outflow", Format: "currency"},
	}
	treasuryBalanceCols = []export.Column{
		{Header: "Período", Key: "period", Format: "text"},
		{Header: "Banco", Key: "bank_name", Format: "text"},
		{Header: "Neto", Key: "net", Format: "currency"},
	}
	paymentMethodCols = []export.Column{
		{Header: "Medio de Pago", Key: "method", Format: "text"},
		{Header: "Cantidad", Key: "count", Format: "number"},
		{Header: "Total", Key: "total", Format: "currency"},
	}
	monthlyTotalsCols = []export.Column{
		{Header: "Período", Key: "period", Format: "text"},
		{Header: "Dirección", Key: "direction", Format: "text"},
		{Header: "Cantidad", Key: "count", Format: "number"},
		{Header: "Total", Key: "total", Format: "currency"},
	}
	taxSummaryCols = []export.Column{
		{Header: "Período", Key: "period", Format: "text"},
		{Header: "Dirección", Key: "direction", Format: "text"},
		{Header: "Neto", Key: "net_total", Format: "currency"},
		{Header: "IVA", Key: "tax_total", Format: "currency"},
	}
	topEntityCols = []export.Column{
		{Header: "Entidad", Key: "entity_name", Format: "text"},
		{Header: "Comprobantes", Key: "invoice_count", Format: "number"},
		{Header: "Total", Key: "total", Format: "currency"},
	}
	customerConcentrationCols = []export.Column{
		{Header: "Entidad", Key: "entity_name", Format: "text"},
		{Header: "Total Ventas", Key: "total_sales", Format: "currency"},
	}
	agingCols = []export.Column{
		{Header: "Entidad", Key: "entity_name", Format: "text"},
		{Header: "Al día", Key: "current_amount", Format: "currency"},
		{Header: "1-30 días", Key: "days_1_30", Format: "currency"},
		{Header: "31-60 días", Key: "days_31_60", Format: "currency"},
		{Header: "61-90 días", Key: "days_61_90", Format: "currency"},
		{Header: "+90 días", Key: "days_over_90", Format: "currency"},
	}
	collectionCols = []export.Column{
		{Header: "Período", Key: "period", Format: "text"},
		{Header: "Facturado", Key: "invoiced", Format: "currency"},
		{Header: "Cobrado", Key: "collected", Format: "currency"},
	}
	stockValuationCols = []export.Column{
		{Header: "Código", Key: "code", Format: "text"},
		{Header: "Nombre", Key: "name", Format: "text"},
		{Header: "Cantidad", Key: "quantity", Format: "number"},
		{Header: "Costo Unit.", Key: "unit_cost", Format: "currency"},
		{Header: "Valor", Key: "stock_value", Format: "currency"},
	}
	stockRotationCols = []export.Column{
		{Header: "Código", Key: "code", Format: "text"},
		{Header: "Nombre", Key: "name", Format: "text"},
		{Header: "Stock Actual", Key: "current_stock", Format: "number"},
		{Header: "Consumido", Key: "consumed_qty", Format: "number"},
		{Header: "Índice Rotación", Key: "rotation_index", Format: "number"},
	}
	belowMinCols = []export.Column{
		{Header: "Código", Key: "code", Format: "text"},
		{Header: "Nombre", Key: "name", Format: "text"},
		{Header: "Stock Mínimo", Key: "min_stock", Format: "number"},
		{Header: "Stock Actual", Key: "current_stock", Format: "number"},
		{Header: "Déficit", Key: "deficit", Format: "number"},
	}
	movementSummaryCols = []export.Column{
		{Header: "Tipo", Key: "movement_type", Format: "text"},
		{Header: "Cantidad Movs.", Key: "count", Format: "number"},
		{Header: "Cantidad Total", Key: "total_quantity", Format: "number"},
		{Header: "Valor Total", Key: "total_value", Format: "currency"},
	}
	prodStatusCols = []export.Column{
		{Header: "Estado", Key: "status", Format: "text"},
		{Header: "Cantidad", Key: "count", Format: "number"},
	}
	prodMonthCols = []export.Column{
		{Header: "Período", Key: "period", Format: "text"},
		{Header: "Total", Key: "total_orders", Format: "number"},
		{Header: "Completadas", Key: "completed", Format: "number"},
	}
	prodEfficiencyCols = []export.Column{
		{Header: "Centro", Key: "center_name", Format: "text"},
		{Header: "Total", Key: "total_orders", Format: "number"},
		{Header: "Completadas", Key: "completed", Format: "number"},
	}
	purchasingMonthlyCols = []export.Column{
		{Header: "Período", Key: "period", Format: "text"},
		{Header: "Cant. OC", Key: "order_count", Format: "number"},
		{Header: "Total", Key: "total", Format: "currency"},
	}
	supplierRankingCols = []export.Column{
		{Header: "Proveedor", Key: "entity_name", Format: "text"},
		{Header: "Cant. OC", Key: "order_count", Format: "number"},
		{Header: "Total Comprado", Key: "total_purchased", Format: "currency"},
		{Header: "Deméritos", Key: "total_demerits", Format: "number"},
	}
	salesMonthlyCols = []export.Column{
		{Header: "Período", Key: "period", Format: "text"},
		{Header: "Cant. Pedidos", Key: "order_count", Format: "number"},
		{Header: "Total", Key: "total", Format: "currency"},
	}
	conversionCols = []export.Column{
		{Header: "Período", Key: "period", Format: "text"},
		{Header: "Total Cotizaciones", Key: "total_quotations", Format: "number"},
		{Header: "Convertidas", Key: "converted", Format: "number"},
	}
	headcountCols = []export.Column{
		{Header: "Departamento", Key: "department", Format: "text"},
		{Header: "Empleados", Key: "headcount", Format: "number"},
	}
	absenceCols = []export.Column{
		{Header: "Período", Key: "period", Format: "text"},
		{Header: "Tipo", Key: "event_type", Format: "text"},
		{Header: "Cantidad", Key: "count", Format: "number"},
	}
	overtimeCols = []export.Column{
		{Header: "Período", Key: "period", Format: "text"},
		{Header: "Eventos", Key: "event_count", Format: "number"},
		{Header: "Horas", Key: "total_hours", Format: "number"},
	}
	ncTypeCols = []export.Column{
		{Header: "Severidad", Key: "severity", Format: "text"},
		{Header: "Estado", Key: "status", Format: "text"},
		{Header: "Cantidad", Key: "count", Format: "number"},
	}
	ncResolutionCols = []export.Column{
		{Header: "Severidad", Key: "severity", Format: "text"},
		{Header: "Total", Key: "total", Format: "number"},
		{Header: "Cerradas", Key: "closed", Format: "number"},
		{Header: "Días Prom.", Key: "avg_days_to_close", Format: "number"},
	}
	woCompletionCols = []export.Column{
		{Header: "Período", Key: "period", Format: "text"},
		{Header: "Total", Key: "total", Format: "number"},
		{Header: "Completadas", Key: "completed", Format: "number"},
	}
	costByAssetCols = []export.Column{
		{Header: "Código", Key: "asset_code", Format: "text"},
		{Header: "Equipo", Key: "asset_name", Format: "text"},
		{Header: "Cant. OT", Key: "wo_count", Format: "number"},
		{Header: "Costo Total", Key: "total_cost", Format: "currency"},
	}
)

// ─── Fetch functions ───────────────────────────────────────────────────────

func (h *Analytics) fetchBalanceEvolution(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsBalanceEvolution(r.Context(), repository.AnalyticsBalanceEvolutionParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"period": r.Period, "account_type": r.AccountType, "total_debit": r.TotalDebit, "total_credit": r.TotalCredit, "net": r.Net}
	}
	return out, nil
}

func (h *Analytics) fetchIncomeExpense(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsIncomeExpense(r.Context(), repository.AnalyticsIncomeExpenseParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"period": r.Period, "income": r.Income, "expense": r.Expense}
	}
	return out, nil
}

func (h *Analytics) fetchCostCenterBreakdown(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsCostCenterBreakdown(r.Context(), repository.AnalyticsCostCenterBreakdownParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"cost_center": r.CostCenter, "account_type": r.AccountType, "total": r.Total}
	}
	return out, nil
}

func (h *Analytics) fetchCashFlow(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsCashFlow(r.Context(), repository.AnalyticsCashFlowParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"period": r.Period, "inflow": r.Inflow, "outflow": r.Outflow}
	}
	return out, nil
}

func (h *Analytics) fetchTreasuryBalance(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsTreasuryBalanceEvolution(r.Context(), repository.AnalyticsTreasuryBalanceEvolutionParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"period": r.Period, "bank_name": r.BankName, "net": r.Net}
	}
	return out, nil
}

func (h *Analytics) fetchPaymentMethods(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsPaymentMethods(r.Context(), repository.AnalyticsPaymentMethodsParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"method": r.Method, "count": r.Count, "total": r.Total}
	}
	return out, nil
}

func (h *Analytics) fetchMonthlyTotals(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsMonthlyTotals(r.Context(), repository.AnalyticsMonthlyTotalsParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"period": r.Period, "direction": r.Direction, "count": r.Count, "total": r.Total}
	}
	return out, nil
}

func (h *Analytics) fetchTaxSummary(r *http.Request, slug string) ([]export.Row, error) {
	periodFrom := r.URL.Query().Get("period_from")
	periodTo := r.URL.Query().Get("period_to")
	if periodFrom == "" {
		periodFrom = time.Now().AddDate(-1, 0, 0).Format("2006-01")
	}
	if periodTo == "" {
		periodTo = time.Now().Format("2006-01")
	}
	rows, err := h.svc.Repo().AnalyticsTaxSummary(r.Context(), repository.AnalyticsTaxSummaryParams{
		TenantID: slug, PeriodFrom: periodFrom, PeriodTo: periodTo,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"period": r.Period, "direction": r.Direction, "net_total": r.NetTotal, "tax_total": r.TaxTotal}
	}
	return out, nil
}

func (h *Analytics) fetchTopCustomers(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsTopCustomers(r.Context(), repository.AnalyticsTopCustomersParams{
		TenantID: slug, DateFrom: from, DateTo: to, TopN: parseTopN(r),
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"entity_name": r.EntityName, "invoice_count": r.InvoiceCount, "total": r.Total}
	}
	return out, nil
}

func (h *Analytics) fetchTopSuppliers(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsTopSuppliers(r.Context(), repository.AnalyticsTopSuppliersParams{
		TenantID: slug, DateFrom: from, DateTo: to, TopN: parseTopN(r),
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"entity_name": r.EntityName, "invoice_count": r.InvoiceCount, "total": r.Total}
	}
	return out, nil
}

func (h *Analytics) fetchAging(r *http.Request, slug string) ([]export.Row, error) {
	direction := r.URL.Query().Get("direction")
	if direction != "receivable" && direction != "payable" {
		direction = "receivable"
	}
	rows, err := h.svc.Repo().AnalyticsAging(r.Context(), repository.AnalyticsAgingParams{
		TenantID: slug, Direction: direction,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"entity_name": r.EntityName, "current_amount": r.CurrentAmount, "days_1_30": r.Days130, "days_31_60": r.Days3160, "days_61_90": r.Days6190, "days_over_90": r.DaysOver90}
	}
	return out, nil
}

func (h *Analytics) fetchCollectionRate(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsCollectionRate(r.Context(), repository.AnalyticsCollectionRateParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"period": r.Period, "invoiced": r.Invoiced, "collected": r.Collected}
	}
	return out, nil
}

func (h *Analytics) fetchStockValuation(r *http.Request, slug string) ([]export.Row, error) {
	rows, err := h.svc.Repo().AnalyticsStockValuation(r.Context(), slug)
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"code": r.Code, "name": r.Name, "quantity": r.Quantity, "unit_cost": r.UnitCost, "stock_value": r.StockValue}
	}
	return out, nil
}

func (h *Analytics) fetchStockRotation(r *http.Request, slug string) ([]export.Row, error) {
	from, _ := parseTsRange(r)
	rows, err := h.svc.Repo().AnalyticsStockRotation(r.Context(), repository.AnalyticsStockRotationParams{
		TenantID: slug, DateFrom: from,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"code": r.Code, "name": r.Name, "current_stock": r.CurrentStock, "consumed_qty": r.ConsumedQty, "rotation_index": r.RotationIndex}
	}
	return out, nil
}

func (h *Analytics) fetchBelowMinimum(r *http.Request, slug string) ([]export.Row, error) {
	rows, err := h.svc.Repo().AnalyticsBelowMinimum(r.Context(), slug)
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"code": r.Code, "name": r.Name, "min_stock": r.MinStock, "current_stock": r.CurrentStock, "deficit": r.Deficit}
	}
	return out, nil
}

func (h *Analytics) fetchMovementSummary(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseTsRange(r)
	rows, err := h.svc.Repo().AnalyticsMovementSummary(r.Context(), repository.AnalyticsMovementSummaryParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"movement_type": r.MovementType, "count": r.Count, "total_quantity": r.TotalQuantity, "total_value": r.TotalValue}
	}
	return out, nil
}

func (h *Analytics) fetchProdByStatus(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsProductionByStatus(r.Context(), repository.AnalyticsProductionByStatusParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"status": r.Status, "count": r.Count}
	}
	return out, nil
}

func (h *Analytics) fetchProdByMonth(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsProductionByMonth(r.Context(), repository.AnalyticsProductionByMonthParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"period": r.Period, "total_orders": r.TotalOrders, "completed": r.Completed}
	}
	return out, nil
}

func (h *Analytics) fetchProdEfficiency(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsProductionEfficiency(r.Context(), repository.AnalyticsProductionEfficiencyParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"center_name": r.CenterName, "total_orders": r.TotalOrders, "completed": r.Completed}
	}
	return out, nil
}

func (h *Analytics) fetchPurchasingMonthly(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsPurchasingByMonth(r.Context(), repository.AnalyticsPurchasingByMonthParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"period": r.Period, "order_count": r.OrderCount, "total": r.Total}
	}
	return out, nil
}

func (h *Analytics) fetchSupplierRanking(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsSupplierRanking(r.Context(), repository.AnalyticsSupplierRankingParams{
		TenantID: slug, DateFrom: from, DateTo: to, TopN: parseTopN(r),
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"entity_name": r.EntityName, "order_count": r.OrderCount, "total_purchased": r.TotalPurchased, "total_demerits": r.TotalDemerits}
	}
	return out, nil
}

func (h *Analytics) fetchSalesMonthly(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsSalesByMonth(r.Context(), repository.AnalyticsSalesByMonthParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"period": r.Period, "order_count": r.OrderCount, "total": r.Total}
	}
	return out, nil
}

func (h *Analytics) fetchQuotationConversion(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsQuotationConversion(r.Context(), repository.AnalyticsQuotationConversionParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"period": r.Period, "total_quotations": r.TotalQuotations, "converted": r.Converted}
	}
	return out, nil
}

func (h *Analytics) fetchCustomerConcentration(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsCustomerConcentration(r.Context(), repository.AnalyticsCustomerConcentrationParams{
		TenantID: slug, DateFrom: from, DateTo: to, TopN: parseTopN(r),
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"entity_name": r.EntityName, "total_sales": r.TotalSales}
	}
	return out, nil
}

func (h *Analytics) fetchHeadcount(r *http.Request, slug string) ([]export.Row, error) {
	rows, err := h.svc.Repo().AnalyticsHeadcount(r.Context(), slug)
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"department": r.Department, "headcount": r.Headcount}
	}
	return out, nil
}

func (h *Analytics) fetchAbsences(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsAbsencesByMonth(r.Context(), repository.AnalyticsAbsencesByMonthParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"period": r.Period, "event_type": r.EventType, "count": r.Count}
	}
	return out, nil
}

func (h *Analytics) fetchOvertime(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsOvertimeSummary(r.Context(), repository.AnalyticsOvertimeSummaryParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"period": r.Period, "event_count": r.EventCount, "total_hours": r.TotalHours}
	}
	return out, nil
}

func (h *Analytics) fetchNCByType(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsNCByType(r.Context(), repository.AnalyticsNCByTypeParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"severity": r.Severity, "status": r.Status, "count": r.Count}
	}
	return out, nil
}

func (h *Analytics) fetchNCResolution(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsNCResolutionTime(r.Context(), repository.AnalyticsNCResolutionTimeParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"severity": r.Severity, "total": r.Total, "closed": r.Closed, "avg_days_to_close": r.AvgDaysToClose}
	}
	return out, nil
}

func (h *Analytics) fetchWOCompletion(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsWOCompletionRate(r.Context(), repository.AnalyticsWOCompletionRateParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"period": r.Period, "total": r.Total, "completed": r.Completed}
	}
	return out, nil
}

func (h *Analytics) fetchCostByAsset(r *http.Request, slug string) ([]export.Row, error) {
	from, to := parseDateRange(r)
	rows, err := h.svc.Repo().AnalyticsCostByAsset(r.Context(), repository.AnalyticsCostByAssetParams{
		TenantID: slug, DateFrom: from, DateTo: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]export.Row, len(rows))
	for i, r := range rows {
		out[i] = export.Row{"asset_code": r.AssetCode, "asset_name": r.AssetName, "wo_count": r.WoCount, "total_cost": r.TotalCost}
	}
	return out, nil
}

// ─── Dashboard KPIs ────────────────────────────────────────────────────────

func (h *Analytics) DashboardKPIs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := tenantSlug(r)

	type kpis struct {
		MonthRevenue       any `json:"month_revenue,omitempty"`
		MonthExpense       any `json:"month_expense,omitempty"`
		CashBalance        any `json:"cash_balance,omitempty"`
		AccountsReceivable any `json:"accounts_receivable,omitempty"`
		AccountsPayable    any `json:"accounts_payable,omitempty"`
		ActiveProdOrders   any `json:"active_prod_orders,omitempty"`
		PendingPurchases   any `json:"pending_purchases,omitempty"`
		OpenQuotations     any `json:"open_quotations,omitempty"`
		StockBelowMin      any `json:"stock_below_min,omitempty"`
		Headcount          any `json:"headcount,omitempty"`
		AbsencesThisMonth  any `json:"absences_this_month,omitempty"`
		OpenNCs            any `json:"open_nonconformities,omitempty"`
		PendingWOs         any `json:"pending_work_orders,omitempty"`
	}

	var result kpis
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errList []string

	sem := make(chan struct{}, 5) // H1: cap concurrent DB queries per request

	run := func(name string, fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			fn()
		}()
	}

	fail := func(name string, err error) {
		slog.Error("dashboard kpi failed", "kpi", name, "error", err)
		mu.Lock()
		errList = append(errList, name) // C1: only return name, not raw DB error
		mu.Unlock()
	}

	repo := h.svc.Repo()

	run("month_revenue", func() {
		if v, err := repo.KPIMonthRevenue(ctx, slug); err == nil {
			mu.Lock(); result.MonthRevenue = v; mu.Unlock()
		} else {
			fail("month_revenue", err)
		}
	})
	run("month_expense", func() {
		if v, err := repo.KPIMonthExpense(ctx, slug); err == nil {
			mu.Lock(); result.MonthExpense = v; mu.Unlock()
		} else {
			fail("month_expense", err)
		}
	})
	run("cash_balance", func() {
		if v, err := repo.KPICashBalance(ctx, slug); err == nil {
			mu.Lock(); result.CashBalance = v; mu.Unlock()
		} else {
			fail("cash_balance", err)
		}
	})
	run("accounts_receivable", func() {
		if v, err := repo.KPIAccountsReceivable(ctx, slug); err == nil {
			mu.Lock(); result.AccountsReceivable = v; mu.Unlock()
		} else {
			fail("accounts_receivable", err)
		}
	})
	run("accounts_payable", func() {
		if v, err := repo.KPIAccountsPayable(ctx, slug); err == nil {
			mu.Lock(); result.AccountsPayable = v; mu.Unlock()
		} else {
			fail("accounts_payable", err)
		}
	})
	run("active_prod_orders", func() {
		if v, err := repo.KPIActiveProdOrders(ctx, slug); err == nil {
			mu.Lock(); result.ActiveProdOrders = v; mu.Unlock()
		} else {
			fail("active_prod_orders", err)
		}
	})
	run("pending_purchases", func() {
		if v, err := repo.KPIPendingPurchases(ctx, slug); err == nil {
			mu.Lock(); result.PendingPurchases = v; mu.Unlock()
		} else {
			fail("pending_purchases", err)
		}
	})
	run("open_quotations", func() {
		if v, err := repo.KPIOpenQuotations(ctx, slug); err == nil {
			mu.Lock(); result.OpenQuotations = v; mu.Unlock()
		} else {
			fail("open_quotations", err)
		}
	})
	run("stock_below_min", func() {
		if v, err := repo.KPIStockBelowMin(ctx, slug); err == nil {
			mu.Lock(); result.StockBelowMin = v; mu.Unlock()
		} else {
			fail("stock_below_min", err)
		}
	})
	run("headcount", func() {
		if v, err := repo.KPIHeadcount(ctx, slug); err == nil {
			mu.Lock(); result.Headcount = v; mu.Unlock()
		} else {
			fail("headcount", err)
		}
	})
	run("absences_this_month", func() {
		if v, err := repo.KPIAbsencesThisMonth(ctx, slug); err == nil {
			mu.Lock(); result.AbsencesThisMonth = v; mu.Unlock()
		} else {
			fail("absences_this_month", err)
		}
	})
	run("open_ncs", func() {
		if v, err := repo.KPIOpenNonconformities(ctx, slug); err == nil {
			mu.Lock(); result.OpenNCs = v; mu.Unlock()
		} else {
			fail("open_ncs", err)
		}
	})
	run("pending_wos", func() {
		if v, err := repo.KPIPendingWorkOrders(ctx, slug); err == nil {
			mu.Lock(); result.PendingWOs = v; mu.Unlock()
		} else {
			fail("pending_wos", err)
		}
	})

	wg.Wait()

	const totalKPIs = 13
	status := http.StatusOK
	if len(errList) == totalKPIs {
		erperrors.WriteError(w, r, erperrors.Wrap(nil, erperrors.CodeInternal, "all dashboard queries failed", http.StatusInternalServerError))
		return
	}
	if len(errList) > 0 {
		status = http.StatusPartialContent
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"kpis":   result,
		"errors": errList,
	})
}
