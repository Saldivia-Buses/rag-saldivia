package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

type Sales struct {
	repo      *repository.Queries
	pool      TxStarter
	audit     *audit.Writer
	publisher *traces.Publisher
}

func NewSales(repo *repository.Queries, pool TxStarter, auditWriter *audit.Writer, publisher *traces.Publisher) *Sales {
	return &Sales{repo: repo, pool: pool, audit: auditWriter, publisher: publisher}
}

func (s *Sales) ListQuotations(ctx context.Context, tenantID, status string, limit, offset int) ([]repository.ListQuotationsRow, error) {
	return s.repo.ListQuotations(ctx, repository.ListQuotationsParams{
		TenantID: tenantID, StatusFilter: status, Limit: int32(limit), Offset: int32(offset),
	})
}

type QuotationDetail struct {
	Quotation repository.ErpQuotation              `json:"quotation"`
	Lines     []repository.ErpQuotationLine        `json:"lines"`
	Options   []repository.ErpQuotationSectionItem `json:"options"`
}

func (s *Sales) GetQuotation(ctx context.Context, id pgtype.UUID, tenantID string) (*QuotationDetail, error) {
	q, err := s.repo.GetQuotation(ctx, repository.GetQuotationParams{ID: id, TenantID: tenantID})
	if err != nil {
		return nil, fmt.Errorf("get quotation: %w", err)
	}
	lines, err := s.repo.ListQuotationLines(ctx, repository.ListQuotationLinesParams{
		QuotationID: id, TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("list lines: %w", err)
	}
	options, err := s.repo.ListQuotationOptions(ctx, repository.ListQuotationOptionsParams{
		TenantID: tenantID, QuotationID: id, Limit: 1000, Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("list options: %w", err)
	}
	return &QuotationDetail{Quotation: q, Lines: lines, Options: options}, nil
}

type CreateQuotationRequest struct {
	TenantID   string
	Number     string
	Date       pgtype.Date
	CustomerID pgtype.UUID
	CurrencyID pgtype.UUID
	ValidUntil pgtype.Date
	Notes      string
	UserID     string
	IP         string
	Lines      []CreateQuotationLineRequest
}

type CreateQuotationLineRequest struct {
	ArticleID   pgtype.UUID
	Description string
	Quantity    string
	UnitPrice   string
}

func (s *Sales) CreateQuotation(ctx context.Context, req CreateQuotationRequest) (*QuotationDetail, error) {
	if req.Number == "" || len(req.Lines) == 0 {
		return nil, fmt.Errorf("number and at least one line required")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := s.repo.WithTx(tx)

	var total pgtype.Numeric
	_ = total.Scan("0")

	q, err := qtx.CreateQuotation(ctx, repository.CreateQuotationParams{
		TenantID: req.TenantID, Number: req.Number, Date: req.Date,
		CustomerID: req.CustomerID, CurrencyID: req.CurrencyID,
		Total: total, ValidUntil: req.ValidUntil, Notes: req.Notes, UserID: req.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("create quotation: %w", err)
	}

	for i, l := range req.Lines {
		_, err := qtx.CreateQuotationLine(ctx, repository.CreateQuotationLineParams{
			TenantID: req.TenantID, QuotationID: q.ID, ArticleID: l.ArticleID,
			Description: l.Description, Quantity: pgNumeric(l.Quantity),
			UnitPrice: pgNumeric(l.UnitPrice), SortOrder: int32(i),
			Metadata: []byte(`{}`),
		})
		if err != nil {
			return nil, fmt.Errorf("create line %d: %w", i, err)
		}
	}

	// Calculate total
	if _, err := tx.Exec(ctx,
		`UPDATE erp_quotations SET total = (
			SELECT COALESCE(SUM(quantity * unit_price), 0) FROM erp_quotation_lines
			WHERE quotation_id = $1 AND tenant_id = $2
		) WHERE id = $1 AND tenant_id = $2`, q.ID, req.TenantID); err != nil {
		return nil, fmt.Errorf("recalculate quotation total: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	s.audit.Write(ctx, audit.Entry{
		TenantID: req.TenantID, UserID: req.UserID,
		Action: "erp.quotation.created", Resource: uuidStr(q.ID),
		Details: map[string]any{"number": req.Number}, IP: req.IP,
	})
	s.publisher.Broadcast(req.TenantID, "erp_sales", map[string]any{
		"action": "quotation_created", "quotation_id": uuidStr(q.ID),
	})

	q, _ = s.repo.GetQuotation(ctx, repository.GetQuotationParams{ID: q.ID, TenantID: req.TenantID})
	lines, _ := s.repo.ListQuotationLines(ctx, repository.ListQuotationLinesParams{
		QuotationID: q.ID, TenantID: req.TenantID,
	})

	slog.Info("quotation created", "id", uuidStr(q.ID), "number", req.Number)
	return &QuotationDetail{Quotation: q, Lines: lines}, nil
}

func (s *Sales) ListOrders(ctx context.Context, tenantID, status, orderType string, limit, offset int) ([]repository.ListOrdersRow, error) {
	return s.repo.ListOrders(ctx, repository.ListOrdersParams{
		TenantID: tenantID, StatusFilter: status, TypeFilter: orderType,
		Limit: int32(limit), Offset: int32(offset),
	})
}

func (s *Sales) CreateOrder(ctx context.Context, tenantID, number string, date pgtype.Date, orderType string, customerID, quotationID pgtype.UUID, total pgtype.Numeric, notes, userID, ip string) (repository.ErpOrder, error) {
	o, err := s.repo.CreateOrder(ctx, repository.CreateOrderParams{
		TenantID: tenantID, Number: number, Date: date, OrderType: orderType,
		CustomerID: customerID, QuotationID: quotationID, Total: total,
		UserID: userID, Notes: notes,
	})
	if err != nil {
		return repository.ErpOrder{}, fmt.Errorf("create order: %w", err)
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.order.created", Resource: uuidStr(o.ID), IP: ip,
	})
	s.publisher.Broadcast(tenantID, "erp_sales", map[string]any{
		"action": "order_created", "order_id": uuidStr(o.ID),
	})
	return o, nil
}

func (s *Sales) UpdateOrderStatus(ctx context.Context, id pgtype.UUID, tenantID, status, userID, ip string) error {
	rows, err := s.repo.UpdateOrderStatus(ctx, repository.UpdateOrderStatusParams{
		ID: id, TenantID: tenantID, Status: status,
	})
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("order not found")
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.order.status_changed", Resource: uuidStr(id),
		Details: map[string]any{"status": status}, IP: ip,
	})
	return nil
}

func (s *Sales) ListPriceLists(ctx context.Context, tenantID string) ([]repository.ErpPriceList, error) {
	return s.repo.ListPriceLists(ctx, tenantID)
}

// PriceListDetail bundles a price list with its items.
type PriceListDetail struct {
	PriceList repository.ErpPriceList            `json:"price_list"`
	Items     []repository.ListPriceListItemsRow `json:"items"`
}

func (s *Sales) GetPriceList(ctx context.Context, id pgtype.UUID, tenantID string) (*PriceListDetail, error) {
	pl, err := s.repo.GetPriceList(ctx, repository.GetPriceListParams{ID: id, TenantID: tenantID})
	if err != nil {
		return nil, fmt.Errorf("get price list: %w", err)
	}
	items, err := s.repo.ListPriceListItems(ctx, repository.ListPriceListItemsParams{
		PriceListID: id, TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("list price list items: %w", err)
	}
	return &PriceListDetail{PriceList: pl, Items: items}, nil
}
