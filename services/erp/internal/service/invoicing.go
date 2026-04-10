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

type Invoicing struct {
	repo      *repository.Queries
	pool      TxStarter
	audit     audit.StrictLogger
	auditLog  *audit.Writer
	publisher *traces.Publisher
}

func NewInvoicing(repo *repository.Queries, pool TxStarter, auditWriter *audit.Writer, publisher *traces.Publisher) *Invoicing {
	return &Invoicing{repo: repo, pool: pool, audit: auditWriter, auditLog: auditWriter, publisher: publisher}
}

var validInvoiceTypes = map[string]bool{
	"invoice_a": true, "invoice_b": true, "invoice_c": true, "invoice_e": true,
	"credit_note": true, "debit_note": true, "delivery_note": true,
}

func (s *Invoicing) ListInvoices(ctx context.Context, tenantID string, typeFilter, dirFilter, statusFilter string, dateFrom, dateTo pgtype.Date, limit, offset int) ([]repository.ListInvoicesRow, error) {
	return s.repo.ListInvoices(ctx, repository.ListInvoicesParams{
		TenantID: tenantID, TypeFilter: typeFilter, DirectionFilter: dirFilter,
		StatusFilter: statusFilter, DateFrom: dateFrom, DateTo: dateTo,
		Limit: int32(limit), Offset: int32(offset),
	})
}

type InvoiceDetail struct {
	Invoice repository.ErpInvoice           `json:"invoice"`
	Lines   []repository.ErpInvoiceLine     `json:"lines"`
}

func (s *Invoicing) GetInvoice(ctx context.Context, id pgtype.UUID, tenantID string) (*InvoiceDetail, error) {
	inv, err := s.repo.GetInvoice(ctx, repository.GetInvoiceParams{ID: id, TenantID: tenantID})
	if err != nil {
		return nil, fmt.Errorf("get invoice: %w", err)
	}
	lines, err := s.repo.ListInvoiceLines(ctx, repository.ListInvoiceLinesParams{
		InvoiceID: id, TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("list lines: %w", err)
	}
	return &InvoiceDetail{Invoice: inv, Lines: lines}, nil
}

type CreateInvoiceRequest struct {
	TenantID    string
	Number      string
	Date        pgtype.Date
	DueDate     pgtype.Date
	InvoiceType string
	Direction   string
	EntityID    pgtype.UUID
	CurrencyID  pgtype.UUID
	OrderID     pgtype.UUID
	UserID      string
	IP          string
	Lines       []CreateInvoiceLineRequest
}

type CreateInvoiceLineRequest struct {
	ArticleID   pgtype.UUID
	Description string
	Quantity    string
	UnitPrice   string
	TaxRate     string
}

func (s *Invoicing) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*InvoiceDetail, error) {
	if req.Number == "" || !validInvoiceTypes[req.InvoiceType] || len(req.Lines) == 0 {
		return nil, fmt.Errorf("number, valid invoice_type, and at least one line required")
	}
	if req.Direction != "issued" && req.Direction != "received" {
		return nil, fmt.Errorf("direction must be issued or received")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := s.repo.WithTx(tx)

	var subtotal, taxAmount, total pgtype.Numeric
	_ = subtotal.Scan("0")
	_ = taxAmount.Scan("0")
	_ = total.Scan("0")

	inv, err := qtx.CreateInvoice(ctx, repository.CreateInvoiceParams{
		TenantID: req.TenantID, Number: req.Number, Date: req.Date,
		DueDate: req.DueDate, InvoiceType: req.InvoiceType,
		Direction: req.Direction, EntityID: req.EntityID,
		CurrencyID: req.CurrencyID, Subtotal: subtotal,
		TaxAmount: taxAmount, Total: total, OrderID: req.OrderID,
		UserID: req.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("create invoice: %w", err)
	}

	for i, l := range req.Lines {
		qty := pgNumeric(l.Quantity)
		price := pgNumeric(l.UnitPrice)
		rate := pgNumeric(l.TaxRate)
		if l.TaxRate == "" {
			_ = rate.Scan("21.00")
		}
		// line_total and tax_amount calculated via SQL would be ideal
		// For now pass 0, recalculate after
		var lineTax, lineTotal pgtype.Numeric
		_ = lineTax.Scan("0")
		_ = lineTotal.Scan("0")

		_, err := qtx.CreateInvoiceLine(ctx, repository.CreateInvoiceLineParams{
			TenantID: req.TenantID, InvoiceID: inv.ID, ArticleID: l.ArticleID,
			Description: l.Description, Quantity: qty, UnitPrice: price,
			TaxRate: rate, TaxAmount: lineTax, LineTotal: lineTotal,
			SortOrder: int32(i),
		})
		if err != nil {
			return nil, fmt.Errorf("create line %d: %w", i, err)
		}
	}

	// Recalculate totals from lines
	if _, err := tx.Exec(ctx, `
		UPDATE erp_invoice_lines SET
			line_total = quantity * unit_price,
			tax_amount = quantity * unit_price * tax_rate / 100
		WHERE invoice_id = $1 AND tenant_id = $2`, inv.ID, req.TenantID); err != nil {
		return nil, fmt.Errorf("recalculate line totals: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE erp_invoices SET
			subtotal = (SELECT COALESCE(SUM(line_total), 0) FROM erp_invoice_lines WHERE invoice_id = $1 AND tenant_id = $2),
			tax_amount = (SELECT COALESCE(SUM(tax_amount), 0) FROM erp_invoice_lines WHERE invoice_id = $1 AND tenant_id = $2),
			total = (SELECT COALESCE(SUM(line_total + tax_amount), 0) FROM erp_invoice_lines WHERE invoice_id = $1 AND tenant_id = $2)
		WHERE id = $1 AND tenant_id = $2`, inv.ID, req.TenantID); err != nil {
		return nil, fmt.Errorf("recalculate invoice totals: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// StrictLogger after commit — avoids phantom audit entries if commit fails
	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID: req.TenantID, UserID: req.UserID,
		Action: "erp.invoice.created", Resource: uuidStr(inv.ID),
		Details: map[string]any{"number": req.Number, "type": req.InvoiceType}, IP: req.IP,
	}); err != nil {
		slog.Error("STRICT audit failed after invoice commit", "error", err, "invoice_id", uuidStr(inv.ID))
	}

	s.publisher.Broadcast(req.TenantID, "erp_invoicing", map[string]any{
		"action": "invoice_created", "invoice_id": uuidStr(inv.ID),
	})

	inv, _ = s.repo.GetInvoice(ctx, repository.GetInvoiceParams{ID: inv.ID, TenantID: req.TenantID})
	lines, _ := s.repo.ListInvoiceLines(ctx, repository.ListInvoiceLinesParams{
		InvoiceID: inv.ID, TenantID: req.TenantID,
	})

	slog.Info("invoice created", "id", uuidStr(inv.ID), "number", req.Number)
	return &InvoiceDetail{Invoice: inv, Lines: lines}, nil
}

func (s *Invoicing) PostInvoice(ctx context.Context, id pgtype.UUID, tenantID, userID, ip string) error {
	rows, err := s.repo.PostInvoice(ctx, repository.PostInvoiceParams{ID: id, TenantID: tenantID})
	if err != nil {
		return fmt.Errorf("post invoice: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("invoice not found or not draft")
	}
	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.invoice.posted", Resource: uuidStr(id), IP: ip,
	}); err != nil {
		return fmt.Errorf("strict audit failed: %w", err)
	}
	s.publisher.Broadcast(tenantID, "erp_invoicing", map[string]any{
		"action": "invoice_posted", "invoice_id": uuidStr(id),
	})
	return nil
}

func (s *Invoicing) GetTaxBook(ctx context.Context, tenantID, period, direction string) ([]repository.GetTaxBookRow, error) {
	return s.repo.GetTaxBook(ctx, repository.GetTaxBookParams{
		TenantID: tenantID, Period: period, DirectionFilter: direction,
	})
}

func (s *Invoicing) ListWithholdings(ctx context.Context, tenantID, typeFilter string, limit, offset int) ([]repository.ListWithholdingsRow, error) {
	return s.repo.ListWithholdings(ctx, repository.ListWithholdingsParams{
		TenantID: tenantID, TypeFilter: typeFilter, Limit: int32(limit), Offset: int32(offset),
	})
}

func (s *Invoicing) CreateWithholding(ctx context.Context, p repository.CreateWithholdingParams, userID, ip string) (repository.ErpWithholding, error) {
	w, err := s.repo.CreateWithholding(ctx, p)
	if err != nil {
		return repository.ErpWithholding{}, fmt.Errorf("create withholding: %w", err)
	}
	s.auditLog.Write(ctx, audit.Entry{
		TenantID: p.TenantID, UserID: userID,
		Action: "erp.withholding.created", Resource: uuidStr(w.ID), IP: ip,
	})
	return w, nil
}
