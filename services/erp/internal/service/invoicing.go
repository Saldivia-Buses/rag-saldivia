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
	Invoice repository.GetInvoiceRow        `json:"invoice"`
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
	defer func() { _ = tx.Rollback(ctx) }()
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

	// StrictLogger before commit — abort if audit fails (financial operation)
	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID: req.TenantID, UserID: req.UserID,
		Action: "erp.invoice.created", Resource: uuidStr(inv.ID),
		Details: map[string]any{"number": req.Number, "type": req.InvoiceType}, IP: req.IP,
	}); err != nil {
		return nil, fmt.Errorf("strict audit failed, aborting: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	s.publisher.Broadcast(req.TenantID, "erp_invoicing", map[string]any{
		"action": "invoice_created", "invoice_id": uuidStr(inv.ID),
	})

	invFresh, _ := s.repo.GetInvoice(ctx, repository.GetInvoiceParams{ID: inv.ID, TenantID: req.TenantID})
	lines, _ := s.repo.ListInvoiceLines(ctx, repository.ListInvoiceLinesParams{
		InvoiceID: inv.ID, TenantID: req.TenantID,
	})

	slog.Info("invoice created", "id", uuidStr(inv.ID), "number", req.Number)
	return &InvoiceDetail{Invoice: invFresh, Lines: lines}, nil
}

// PostInvoice posts a draft invoice: changes status, generates tax entries (libro IVA),
// all in a single transaction.
func (s *Invoicing) PostInvoice(ctx context.Context, id pgtype.UUID, tenantID, userID, ip string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := s.repo.WithTx(tx)

	// 1. Flip status to posted
	rows, err := qtx.PostInvoice(ctx, repository.PostInvoiceParams{ID: id, TenantID: tenantID})
	if err != nil {
		return fmt.Errorf("post invoice: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("invoice not found or not draft")
	}

	// 2. Fetch invoice + lines to generate tax entries
	inv, err := qtx.GetInvoice(ctx, repository.GetInvoiceParams{ID: id, TenantID: tenantID})
	if err != nil {
		return fmt.Errorf("get invoice for tax gen: %w", err)
	}

	lines, err := qtx.ListInvoiceLines(ctx, repository.ListInvoiceLinesParams{
		InvoiceID: id, TenantID: tenantID,
	})
	if err != nil {
		return fmt.Errorf("list lines for tax gen: %w", err)
	}

	// 3. Generate tax entries grouped by tax_rate
	// Period from invoice date (YYYY-MM)
	var period string
	if inv.Date.Valid {
		t := inv.Date.Time
		period = fmt.Sprintf("%04d-%02d", t.Year(), t.Month())
	}

	direction := "sales"
	if inv.Direction == "received" {
		direction = "purchases"
	}

	for _, line := range lines {
		_, err := qtx.CreateTaxEntry(ctx, repository.CreateTaxEntryParams{
			TenantID:  tenantID,
			InvoiceID: id,
			Period:    period,
			Direction: direction,
			NetAmount: line.LineTotal,
			TaxRate:   line.TaxRate,
			TaxAmount: line.TaxAmount,
		})
		if err != nil {
			return fmt.Errorf("create tax entry: %w", err)
		}
	}

	// StrictLogger before commit — fail-closed (pattern P7)
	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.invoice.posted", Resource: uuidStr(id),
		Details: map[string]any{"period": period, "tax_entries": len(lines)}, IP: ip,
	}); err != nil {
		return fmt.Errorf("strict audit failed, aborting: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit post: %w", err)
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

// ============================================================
// Cascade void (Plan 18 Fase 2)
// ============================================================

// VoidResult holds the result of a cascade void operation.
type VoidResult struct {
	ReversalEntryID      pgtype.UUID `json:"reversal_entry_id,omitempty"`
	TaxEntriesReversed   int         `json:"tax_entries_reversed"`
	AcctMovementsReversed int        `json:"acct_movements_reversed"`
	StockMovementsReversed int       `json:"stock_movements_reversed"`
}

// VoidPreview returns what voiding an invoice would do.
func (s *Invoicing) VoidPreview(ctx context.Context, id pgtype.UUID, tenantID string) (*VoidPreviewResult, error) {
	inv, err := s.repo.GetInvoice(ctx, repository.GetInvoiceParams{ID: id, TenantID: tenantID})
	if err != nil {
		return nil, fmt.Errorf("get invoice: %w", err)
	}
	if inv.Status != "posted" && inv.Status != "paid" {
		return nil, fmt.Errorf("solo se pueden anular facturas posted/paid (actual: %s)", inv.Status)
	}
	// Block if invoice has CAE (needs Plan 19 for NCA workflow)
	if inv.AfipCae.Valid && inv.AfipCae.String != "" {
		return nil, fmt.Errorf("factura con CAE no se puede anular sin Plan 19 (AFIP NCA)")
	}

	taxEntries, _ := s.repo.ListTaxEntriesByInvoice(ctx, repository.ListTaxEntriesByInvoiceParams{
		TenantID: tenantID, InvoiceID: id,
	})
	acctMovs, _ := s.repo.ListAccountMovementsByInvoice(ctx, repository.ListAccountMovementsByInvoiceParams{
		TenantID: tenantID, InvoiceID: id,
	})
	stockMovs, _ := s.repo.ListStockMovementsByRef(ctx, repository.ListStockMovementsByRefParams{
		TenantID: tenantID, ReferenceType: pgText("invoice"), ReferenceID: id,
	})

	return &VoidPreviewResult{
		Invoice:           inv,
		HasJournalEntry:   inv.JournalEntryID.Valid,
		TaxEntryCount:     len(taxEntries),
		AcctMovementCount: len(acctMovs),
		StockMovementCount: len(stockMovs),
	}, nil
}

// VoidPreviewResult holds the preview data.
type VoidPreviewResult struct {
	Invoice            repository.GetInvoiceRow `json:"invoice"`
	HasJournalEntry    bool                  `json:"has_journal_entry"`
	TaxEntryCount      int                   `json:"tax_entry_count"`
	AcctMovementCount  int                   `json:"acct_movement_count"`
	StockMovementCount int                   `json:"stock_movement_count"`
}

// VoidInvoice performs cascade void: reverses journal entry, IVA entries,
// account movements, and stock movements in a single atomic transaction.
func (s *Invoicing) VoidInvoice(ctx context.Context, id pgtype.UUID, tenantID, reason, userID, ip string) (*VoidResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := s.repo.WithTx(tx)

	inv, err := qtx.GetInvoice(ctx, repository.GetInvoiceParams{ID: id, TenantID: tenantID})
	if err != nil {
		return nil, fmt.Errorf("get invoice: %w", err)
	}
	if inv.Status != "posted" && inv.Status != "paid" {
		return nil, fmt.Errorf("solo se pueden anular facturas posted/paid")
	}
	// Block CAE invoices until Plan 19
	if inv.AfipCae.Valid && inv.AfipCae.String != "" {
		return nil, fmt.Errorf("factura con CAE requiere Nota de Crédito AFIP (Plan 19)")
	}

	result := &VoidResult{}

	// 1. Reverse journal entry (if exists)
	if inv.JournalEntryID.Valid {
		reversalEntry, err := s.createReversalEntry(ctx, qtx, tenantID, inv.JournalEntryID, inv.Date, userID)
		if err != nil {
			return nil, fmt.Errorf("create reversal entry: %w", err)
		}
		result.ReversalEntryID = reversalEntry.ID

		// Mark original as reversed
		_, err = qtx.MarkEntryReversed(ctx, repository.MarkEntryReversedParams{
			ID: inv.JournalEntryID, TenantID: tenantID, ReversedBy: reversalEntry.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("mark entry reversed: %w", err)
		}
	}

	// 2. Reverse tax entries (create negated copies)
	taxEntries, err := qtx.ListTaxEntriesByInvoice(ctx, repository.ListTaxEntriesByInvoiceParams{
		TenantID: tenantID, InvoiceID: id,
	})
	if err != nil {
		return nil, fmt.Errorf("list tax entries: %w", err)
	}
	for _, te := range taxEntries {
		_, err := qtx.CreateTaxEntry(ctx, repository.CreateTaxEntryParams{
			TenantID:  tenantID,
			InvoiceID: id,
			Period:    te.Period,
			Direction: te.Direction,
			NetAmount: negateNumeric(te.NetAmount),
			TaxRate:   te.TaxRate,
			TaxAmount: negateNumeric(te.TaxAmount),
		})
		if err != nil {
			return nil, fmt.Errorf("reverse tax entry: %w", err)
		}
	}
	result.TaxEntriesReversed = len(taxEntries)

	// 3. Reverse account movements
	acctMovs, err := qtx.ListAccountMovementsByInvoice(ctx, repository.ListAccountMovementsByInvoiceParams{
		TenantID: tenantID, InvoiceID: id,
	})
	if err != nil {
		return nil, fmt.Errorf("list account movements: %w", err)
	}
	for _, am := range acctMovs {
		_, err := qtx.CreateAccountMovement(ctx, repository.CreateAccountMovementParams{
			TenantID:       tenantID,
			EntityID:       am.EntityID,
			Date:           am.Date,
			MovementType:   "reversal",
			Direction:      am.Direction,
			Amount:         negateNumeric(am.Amount),
			Balance:        zeroNumeric(),
			InvoiceID:      am.InvoiceID,
			JournalEntryID: result.ReversalEntryID,
			Notes:          "Reversa por anulación de factura",
			UserID:         userID,
		})
		if err != nil {
			return nil, fmt.Errorf("reverse account movement: %w", err)
		}
	}
	result.AcctMovementsReversed = len(acctMovs)

	// 4. Reverse stock movements (if any)
	stockMovs, err := qtx.ListStockMovementsByRef(ctx, repository.ListStockMovementsByRefParams{
		TenantID: tenantID, ReferenceType: pgText("invoice"), ReferenceID: id,
	})
	if err != nil {
		return nil, fmt.Errorf("list stock movements: %w", err)
	}
	for _, sm := range stockMovs {
		reverseType := "in"
		if sm.MovementType == "in" {
			reverseType = "out"
		}
		newMov, err := qtx.CreateStockMovement(ctx, repository.CreateStockMovementParams{
			TenantID:      tenantID,
			ArticleID:     sm.ArticleID,
			WarehouseID:   sm.WarehouseID,
			MovementType:  reverseType,
			Quantity:      sm.Quantity,
			UnitCost:      sm.UnitCost,
			ReferenceType: pgText("void"),
			ReferenceID:   id,
			UserID:        userID,
			Notes:         "Reversa por anulación de factura",
		})
		if err != nil {
			return nil, fmt.Errorf("reverse stock movement: %w", err)
		}
		// Update stock levels
		qtyDelta := sm.Quantity
		if reverseType == "out" {
			qtyDelta = negateNumeric(sm.Quantity)
		}
		_ = qtx.UpsertStockLevel(ctx, repository.UpsertStockLevelParams{
			TenantID: tenantID, ArticleID: sm.ArticleID,
			WarehouseID: sm.WarehouseID, Quantity: qtyDelta,
		})
		_ = newMov // suppress unused
	}
	result.StockMovementsReversed = len(stockMovs)

	// 5. Mark invoice as cancelled
	rows, err := qtx.VoidInvoice(ctx, repository.VoidInvoiceParams{
		ID: id, TenantID: tenantID,
		VoidReason: pgText(reason),
	})
	if err != nil {
		return nil, fmt.Errorf("void invoice: %w", err)
	}
	if rows == 0 {
		return nil, fmt.Errorf("invoice void failed (concurrent modification?)")
	}

	// StrictLogger
	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.invoice.void_cascade", Resource: uuidStr(id),
		Details: map[string]any{
			"reason":             reason,
			"reversal_entry":     uuidStr(result.ReversalEntryID),
			"tax_reversed":       result.TaxEntriesReversed,
			"acct_mov_reversed":  result.AcctMovementsReversed,
			"stock_mov_reversed": result.StockMovementsReversed,
		}, IP: ip,
	}); err != nil {
		return nil, fmt.Errorf("strict audit failed, aborting: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit void: %w", err)
	}

	s.publisher.Broadcast(tenantID, "erp_invoicing", map[string]any{
		"action": "invoice_voided", "invoice_id": uuidStr(id),
	})

	return result, nil
}

// createReversalEntry creates a reversal journal entry with inverted debit/credit lines.
func (s *Invoicing) createReversalEntry(ctx context.Context, qtx *repository.Queries, tenantID string, originalEntryID pgtype.UUID, date pgtype.Date, userID string) (repository.CreateJournalEntryRow, error) {
	original, err := qtx.GetJournalEntry(ctx, repository.GetJournalEntryParams{
		ID: originalEntryID, TenantID: tenantID,
	})
	if err != nil {
		return repository.CreateJournalEntryRow{}, fmt.Errorf("get original entry: %w", err)
	}

	lines, err := qtx.ListJournalLines(ctx, repository.ListJournalLinesParams{
		EntryID: originalEntryID, TenantID: tenantID,
	})
	if err != nil {
		return repository.CreateJournalEntryRow{}, fmt.Errorf("list original lines: %w", err)
	}

	// Create reversal entry
	entry, err := qtx.CreateJournalEntry(ctx, repository.CreateJournalEntryParams{
		TenantID:     tenantID,
		Number:       "REV-" + original.Number,
		Date:         date,
		FiscalYearID: original.FiscalYearID,
		Concept:      "Reversa: " + original.Concept,
		EntryType:    "reversal",
		ReferenceType: pgText("reversal"),
		ReferenceID:  originalEntryID,
		UserID:       userID,
	})
	if err != nil {
		return repository.CreateJournalEntryRow{}, fmt.Errorf("create reversal entry: %w", err)
	}

	// Invert debit/credit on each line
	for i, l := range lines {
		_, err := qtx.CreateJournalLine(ctx, repository.CreateJournalLineParams{
			TenantID:     tenantID,
			EntryID:      entry.ID,
			AccountID:    l.AccountID,
			CostCenterID: l.CostCenterID,
			EntryDate:    date,
			Debit:        l.Credit, // swap
			Credit:       l.Debit,  // swap
			Description:  "Reversa: " + l.Description,
			SortOrder:    int32(i),
		})
		if err != nil {
			return repository.CreateJournalEntryRow{}, fmt.Errorf("create reversal line %d: %w", i, err)
		}
	}

	// Post immediately
	_, err = qtx.PostJournalEntry(ctx, repository.PostJournalEntryParams{
		ID: entry.ID, TenantID: tenantID,
	})
	if err != nil {
		return repository.CreateJournalEntryRow{}, fmt.Errorf("post reversal entry: %w", err)
	}

	return entry, nil
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
