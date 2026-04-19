package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

// Treasury handles treasury business logic.
type Treasury struct {
	repo      *repository.Queries
	pool      TxStarter
	audit     audit.StrictLogger
	auditLog  *audit.Writer
	publisher *traces.Publisher
}

// NewTreasury creates a treasury service.
func NewTreasury(repo *repository.Queries, pool TxStarter, auditWriter *audit.Writer, publisher *traces.Publisher) *Treasury {
	return &Treasury{repo: repo, pool: pool, audit: auditWriter, auditLog: auditWriter, publisher: publisher}
}

var validMovementTypesT = map[string]bool{
	"cash_in": true, "cash_out": true, "bank_deposit": true,
	"bank_withdrawal": true, "check_issued": true, "check_received": true, "transfer": true,
}

func (s *Treasury) ListBankAccounts(ctx context.Context, tenantID string, activeOnly bool) ([]repository.ErpBankAccount, error) {
	return s.repo.ListBankAccounts(ctx, repository.ListBankAccountsParams{TenantID: tenantID, ActiveOnly: activeOnly})
}

func (s *Treasury) CreateBankAccount(ctx context.Context, p repository.CreateBankAccountParams, userID, ip string) (repository.ErpBankAccount, error) {
	ba, err := s.repo.CreateBankAccount(ctx, p)
	if err != nil {
		return repository.ErpBankAccount{}, fmt.Errorf("create bank account: %w", err)
	}
	s.auditLog.Write(ctx, audit.Entry{
		TenantID: p.TenantID, UserID: userID,
		Action: "erp.bank_account.created", Resource: uuidStr(ba.ID), IP: ip,
	})
	return ba, nil
}

func (s *Treasury) ListCashRegisters(ctx context.Context, tenantID string) ([]repository.ErpCashRegister, error) {
	return s.repo.ListCashRegisters(ctx, tenantID)
}

func (s *Treasury) CreateCashRegister(ctx context.Context, tenantID, name string, accountID pgtype.UUID, userID, ip string) (repository.ErpCashRegister, error) {
	cr, err := s.repo.CreateCashRegister(ctx, repository.CreateCashRegisterParams{
		TenantID: tenantID, Name: name, AccountID: accountID,
	})
	if err != nil {
		return repository.ErpCashRegister{}, fmt.Errorf("create cash register: %w", err)
	}
	s.auditLog.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.cash_register.created", Resource: uuidStr(cr.ID), IP: ip,
	})
	return cr, nil
}

func (s *Treasury) ListMovements(ctx context.Context, tenantID string, dateFrom, dateTo pgtype.Date, typeFilter string, limit, offset int) ([]repository.ListTreasuryMovementsRow, error) {
	return s.repo.ListTreasuryMovements(ctx, repository.ListTreasuryMovementsParams{
		TenantID: tenantID, DateFrom: dateFrom, DateTo: dateTo,
		TypeFilter: typeFilter, Limit: int32(limit), Offset: int32(offset),
	})
}

// CreateMovementRequest holds data for creating a treasury movement.
type CreateTreasuryMovementRequest struct {
	repository.CreateTreasuryMovementParams
	UserIDVal string
	IP        string
}

func (s *Treasury) CreateMovement(ctx context.Context, req CreateTreasuryMovementRequest) (repository.CreateTreasuryMovementRow, error) {
	if !validMovementTypesT[req.MovementType] {
		return repository.CreateTreasuryMovementRow{}, fmt.Errorf("invalid movement type: %s", req.MovementType)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return repository.CreateTreasuryMovementRow{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := s.repo.WithTx(tx)

	mov, err := qtx.CreateTreasuryMovement(ctx, req.CreateTreasuryMovementParams)
	if err != nil {
		return repository.CreateTreasuryMovementRow{}, fmt.Errorf("create movement: %w", err)
	}

	// StrictLogger — abort on failure (pattern P7)
	idStr := uuidStr(mov.ID)
	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID: req.TenantID, UserID: req.UserIDVal,
		Action: "erp.treasury.movement", Resource: idStr,
		Details: map[string]any{"type": req.MovementType}, IP: req.IP,
	}); err != nil {
		return repository.CreateTreasuryMovementRow{}, fmt.Errorf("strict audit failed, aborting: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return repository.CreateTreasuryMovementRow{}, fmt.Errorf("commit movement: %w", err)
	}

	s.publisher.Broadcast(req.TenantID, "erp_treasury", map[string]any{
		"action": "movement_created", "movement_id": idStr,
	})
	return mov, nil
}

func (s *Treasury) ListChecks(ctx context.Context, tenantID, direction, status string) ([]repository.ErpCheck, error) {
	return s.repo.ListChecks(ctx, repository.ListChecksParams{
		TenantID: tenantID, DirectionFilter: direction, StatusFilter: status,
	})
}

func (s *Treasury) CreateCheck(ctx context.Context, p repository.CreateCheckParams, userID, ip string) (repository.ErpCheck, error) {
	chk, err := s.repo.CreateCheck(ctx, p)
	if err != nil {
		return repository.ErpCheck{}, fmt.Errorf("create check: %w", err)
	}
	s.auditLog.Write(ctx, audit.Entry{
		TenantID: p.TenantID, UserID: userID,
		Action: "erp.check.created", Resource: uuidStr(chk.ID), IP: ip,
	})
	return chk, nil
}

var validCheckTransitions = map[string][]string{
	"in_portfolio": {"deposited", "endorsed", "rejected"},
	"deposited":    {"cashed", "rejected"},
	"endorsed":     {"cashed", "rejected"},
}

func (s *Treasury) UpdateCheckStatus(ctx context.Context, id pgtype.UUID, tenantID, newStatus, userID, ip string) error {
	// Fetch current check by ID to validate transition
	chk, err := s.repo.GetCheck(ctx, repository.GetCheckParams{ID: id, TenantID: tenantID})
	if err != nil {
		return fmt.Errorf("check not found")
	}
	currentStatus := chk.Status

	// Validate transition
	allowed := validCheckTransitions[currentStatus]
	valid := false
	for _, s := range allowed {
		if s == newStatus {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid transition from %s to %s", currentStatus, newStatus)
	}

	rows, err := s.repo.UpdateCheckStatus(ctx, repository.UpdateCheckStatusParams{
		ID: id, TenantID: tenantID, Status: newStatus,
	})
	if err != nil {
		return fmt.Errorf("update check: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("check not found")
	}
	s.auditLog.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.check.status_changed", Resource: uuidStr(id),
		Details: map[string]any{"status": newStatus}, IP: ip,
	})
	return nil
}

func (s *Treasury) GetBalance(ctx context.Context, tenantID string) ([]repository.GetTreasuryBalanceRow, error) {
	return s.repo.GetTreasuryBalance(ctx, tenantID)
}

func (s *Treasury) ListCashCounts(ctx context.Context, tenantID string, limit, offset int) ([]repository.ErpCashCount, error) {
	return s.repo.ListCashCounts(ctx, repository.ListCashCountsParams{
		TenantID: tenantID, Limit: int32(limit), Offset: int32(offset),
	})
}

func (s *Treasury) CreateCashCount(ctx context.Context, p repository.CreateCashCountParams, ip string) (repository.ErpCashCount, error) {
	cc, err := s.repo.CreateCashCount(ctx, p)
	if err != nil {
		return repository.ErpCashCount{}, fmt.Errorf("create cash count: %w", err)
	}
	s.auditLog.Write(ctx, audit.Entry{
		TenantID: p.TenantID, UserID: p.UserID,
		Action: "erp.cash_count.created", Resource: uuidStr(cc.ID), IP: ip,
	})
	return cc, nil
}

// ============================================================
// Reconciliation (Plan 18 Fase 1)
// ============================================================

func (s *Treasury) ListReconciliations(ctx context.Context, tenantID string) ([]repository.ListReconciliationsRow, error) {
	return s.repo.ListReconciliations(ctx, tenantID)
}

func (s *Treasury) CreateReconciliation(ctx context.Context, tenantID string, bankAccountID pgtype.UUID, period string, statementBalance, bookBalance string, userID, ip string) (repository.ErpBankReconciliation, error) {
	recon, err := s.repo.CreateReconciliation(ctx, repository.CreateReconciliationParams{
		TenantID: tenantID, BankAccountID: bankAccountID, Period: period,
		StatementBalance: pgNumeric(statementBalance), BookBalance: pgNumeric(bookBalance),
		UserID: userID,
	})
	if err != nil {
		return repository.ErpBankReconciliation{}, fmt.Errorf("create reconciliation: %w", err)
	}
	s.auditLog.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.reconciliation.created", Resource: uuidStr(recon.ID),
		Details: map[string]any{"period": period}, IP: ip,
	})
	return recon, nil
}

// ReconciliationDetail bundles a reconciliation with its statement lines.
type ReconciliationDetail struct {
	Reconciliation repository.GetReconciliationRow `json:"reconciliation"`
	Lines          []repository.ErpBankStatementLine `json:"lines"`
}

func (s *Treasury) GetReconciliation(ctx context.Context, tenantID string, id pgtype.UUID) (*ReconciliationDetail, error) {
	recon, err := s.repo.GetReconciliation(ctx, repository.GetReconciliationParams{ID: id, TenantID: tenantID})
	if err != nil {
		return nil, fmt.Errorf("get reconciliation: %w", err)
	}
	lines, err := s.repo.ListStatementLines(ctx, repository.ListStatementLinesParams{
		TenantID: tenantID, ReconciliationID: id,
	})
	if err != nil {
		return nil, fmt.Errorf("list statement lines: %w", err)
	}
	return &ReconciliationDetail{Reconciliation: recon, Lines: lines}, nil
}

// ImportStatementLines imports bank statement lines into a reconciliation.
type StatementLineInput struct {
	Date        string `json:"date"`
	Description string `json:"description"`
	Amount      string `json:"amount"`
	Reference   string `json:"reference"`
}

func (s *Treasury) ImportStatementLines(ctx context.Context, tenantID string, reconID pgtype.UUID, lines []StatementLineInput) (int, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := s.repo.WithTx(tx)

	count := 0
	for _, l := range lines {
		var d pgtype.Date
		_ = d.Scan(l.Date)
		_, err := qtx.CreateStatementLine(ctx, repository.CreateStatementLineParams{
			TenantID:         tenantID,
			ReconciliationID: reconID,
			Date:             d,
			Description:      l.Description,
			Amount:           pgNumeric(l.Amount),
			Reference:        l.Reference,
		})
		if err != nil {
			return 0, fmt.Errorf("create line %d: %w", count, err)
		}
		count++
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit import: %w", err)
	}
	return count, nil
}

// AutoMatchResult holds the result of auto-matching.
type AutoMatchResult struct {
	Matched   int `json:"matched"`
	Unmatched int `json:"unmatched"`
}

// AutoMatch matches statement lines against treasury movements by amount + date (±2 days).
func (s *Treasury) AutoMatch(ctx context.Context, tenantID string, reconID pgtype.UUID) (*AutoMatchResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := s.repo.WithTx(tx)

	recon, err := qtx.GetReconciliation(ctx, repository.GetReconciliationParams{ID: reconID, TenantID: tenantID})
	if err != nil {
		return nil, fmt.Errorf("get reconciliation: %w", err)
	}

	lines, err := qtx.ListUnmatchedStatementLines(ctx, repository.ListUnmatchedStatementLinesParams{
		TenantID: tenantID, ReconciliationID: reconID,
	})
	if err != nil {
		return nil, fmt.Errorf("list unmatched lines: %w", err)
	}

	movements, err := qtx.ListUnreconciledMovements(ctx, repository.ListUnreconciledMovementsParams{
		TenantID: tenantID, BankAccountID: recon.BankAccountID,
		Period: recon.Period,
	})
	if err != nil {
		return nil, fmt.Errorf("list unreconciled movements: %w", err)
	}

	// Track which movements have been matched to avoid double-matching
	movMatched := make(map[string]bool) // movement UUID string → matched

	matched := 0
	for _, line := range lines {
		for _, mov := range movements {
			movKey := uuidStr(mov.ID)
			if movMatched[movKey] {
				continue
			}
			if amountsMatch(line.Amount, mov.Amount, mov.MovementType) && datesClose(line.Date, mov.Date, 2) {
				// Match the statement line to the movement
				if _, err := qtx.MatchStatementLine(ctx, repository.MatchStatementLineParams{
					ID: line.ID, TenantID: tenantID, MovementID: mov.ID,
				}); err != nil {
					continue // skip this match, try next movement
				}
				if _, err := qtx.MarkMovementReconciled(ctx, repository.MarkMovementReconciledParams{
					ID: mov.ID, TenantID: tenantID, ReconciliationID: reconID,
				}); err != nil {
					continue
				}
				movMatched[movKey] = true
				matched++
				break
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit auto-match: %w", err)
	}

	s.publisher.Broadcast(tenantID, "erp_treasury", map[string]any{
		"action": "reconciliation_auto_matched", "reconciliation_id": uuidStr(reconID),
	})

	return &AutoMatchResult{Matched: matched, Unmatched: len(lines) - matched}, nil
}

// MatchManual manually matches a statement line to a treasury movement.
func (s *Treasury) MatchManual(ctx context.Context, tenantID string, reconID, lineID, movementID pgtype.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := s.repo.WithTx(tx)

	rows, err := qtx.MatchStatementLine(ctx, repository.MatchStatementLineParams{
		ID: lineID, TenantID: tenantID, MovementID: movementID,
	})
	if err != nil {
		return fmt.Errorf("match line: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("line not found or already matched")
	}

	_, err = qtx.MarkMovementReconciled(ctx, repository.MarkMovementReconciledParams{
		ID: movementID, TenantID: tenantID, ReconciliationID: reconID,
	})
	if err != nil {
		return fmt.Errorf("mark reconciled: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit manual match: %w", err)
	}
	return nil
}

// ConfirmReconciliation confirms a reconciliation within a TX (atomic with audit).
func (s *Treasury) ConfirmReconciliation(ctx context.Context, tenantID string, reconID pgtype.UUID, userID, ip string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := s.repo.WithTx(tx)

	rows, err := qtx.ConfirmReconciliation(ctx, repository.ConfirmReconciliationParams{
		ID: reconID, TenantID: tenantID,
	})
	if err != nil {
		return fmt.Errorf("confirm reconciliation: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("reconciliation not found or already confirmed")
	}

	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.reconciliation.confirmed", Resource: uuidStr(reconID), IP: ip,
	}); err != nil {
		return fmt.Errorf("strict audit failed, aborting: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit confirm: %w", err)
	}

	s.publisher.Broadcast(tenantID, "erp_treasury", map[string]any{
		"action": "reconciliation_confirmed", "reconciliation_id": uuidStr(reconID),
	})
	return nil
}

// ============================================================
// Receipts (Plan 18 Fase 4)
// ============================================================

// ReceiptInput holds data for creating a receipt.
type ReceiptInput struct {
	ReceiptType  string                `json:"receipt_type"`  // "collection" | "payment"
	EntityID     string                `json:"entity_id"`
	Date         string                `json:"date"`
	Notes        string                `json:"notes"`
	Payments     []ReceiptPaymentInput `json:"payments"`
	Allocations  []ReceiptAllocInput   `json:"allocations"`
	Withholdings []ReceiptWHInput      `json:"withholdings"`
}

type ReceiptPaymentInput struct {
	PaymentMethod string  `json:"payment_method"` // 'cash','check','transfer','echeq'
	Amount        string  `json:"amount"`
	BankAccountID *string `json:"bank_account_id,omitempty"`
	CheckID       *string `json:"check_id,omitempty"`
	Notes         string  `json:"notes"`
}

type ReceiptAllocInput struct {
	InvoiceID string `json:"invoice_id"`
	Amount    string `json:"amount"`
}

type ReceiptWHInput struct {
	Type           string `json:"type"`
	EntityID       string `json:"entity_id"`
	Rate           string `json:"rate"`
	BaseAmount     string `json:"base_amount"`
	Amount         string `json:"amount"`
	CertificateNum string `json:"certificate_num"`
	Date           string `json:"date"`
}

// ReceiptDetail bundles a receipt with its payments, allocations, and withholdings.
type ReceiptDetail struct {
	Receipt     repository.GetReceiptRow             `json:"receipt"`
	Payments    []repository.ErpReceiptPayment       `json:"payments"`
	Allocations []repository.ListReceiptAllocationsRow `json:"allocations"`
}

func (s *Treasury) ListReceipts(ctx context.Context, tenantID, typeFilter string, dateFrom, dateTo pgtype.Date, limit, offset int) ([]repository.ListReceiptsRow, error) {
	return s.repo.ListReceipts(ctx, repository.ListReceiptsParams{
		TenantID: tenantID, TypeFilter: typeFilter,
		DateFrom: dateFrom, DateTo: dateTo,
		Limit: int32(limit), Offset: int32(offset),
	})
}

func (s *Treasury) GetReceipt(ctx context.Context, tenantID string, id pgtype.UUID) (*ReceiptDetail, error) {
	receipt, err := s.repo.GetReceipt(ctx, repository.GetReceiptParams{ID: id, TenantID: tenantID})
	if err != nil {
		return nil, fmt.Errorf("get receipt: %w", err)
	}
	payments, err := s.repo.ListReceiptPayments(ctx, repository.ListReceiptPaymentsParams{
		TenantID: tenantID, ReceiptID: id,
	})
	if err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}
	allocs, err := s.repo.ListReceiptAllocations(ctx, repository.ListReceiptAllocationsParams{
		TenantID: tenantID, ReceiptID: id,
	})
	if err != nil {
		return nil, fmt.Errorf("list allocations: %w", err)
	}
	return &ReceiptDetail{Receipt: receipt, Payments: payments, Allocations: allocs}, nil
}

// ErrReceiptUnbalanced is returned when sum(payments)+sum(withholdings) != sum(allocations).
var ErrReceiptUnbalanced = fmt.Errorf("receipt totals don't balance: sum(payments) + sum(withholdings) must equal sum(allocations)")

// CreateReceipt creates a receipt with payments, allocations, and withholdings in one TX.
// Generates treasury movements for each payment + account movements for each allocation.
func (s *Treasury) CreateReceipt(ctx context.Context, tenantID string, inp ReceiptInput, userID, ip string) (*ReceiptDetail, error) {
	if inp.ReceiptType != "collection" && inp.ReceiptType != "payment" {
		return nil, fmt.Errorf("receipt_type must be 'collection' or 'payment'")
	}
	if len(inp.Payments) == 0 || len(inp.Allocations) == 0 {
		return nil, fmt.Errorf("at least one payment and one allocation required")
	}

	// Validate balance: sum(payments) + sum(withholdings) == sum(allocations)
	var payTotal, whTotal, allocTotal float64
	for _, p := range inp.Payments {
		payTotal += parseFloatApprox(p.Amount)
	}
	for _, w := range inp.Withholdings {
		whTotal += parseFloatApprox(w.Amount)
	}
	for _, a := range inp.Allocations {
		allocTotal += parseFloatApprox(a.Amount)
	}
	diff := (payTotal + whTotal) - allocTotal
	if diff > 0.01 || diff < -0.01 {
		return nil, fmt.Errorf("%w: payments=%.2f withholdings=%.2f allocations=%.2f diff=%.2f",
			ErrReceiptUnbalanced, payTotal, whTotal, allocTotal, diff)
	}

	entityID, err := parseUUIDStr(inp.EntityID)
	if err != nil {
		return nil, fmt.Errorf("invalid entity_id: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := s.repo.WithTx(tx)

	// Get next receipt number
	nextNum, err := qtx.GetNextReceiptNumber(ctx, repository.GetNextReceiptNumberParams{
		TenantID: tenantID, ReceiptType: inp.ReceiptType,
	})
	if err != nil {
		return nil, fmt.Errorf("get next receipt number: %w", err)
	}
	number := fmt.Sprintf("REC-%s-%06d", inp.ReceiptType[:3], nextNum)

	var d pgtype.Date
	_ = d.Scan(inp.Date)

	// 1. Create receipt
	receipt, err := qtx.CreateReceipt(ctx, repository.CreateReceiptParams{
		TenantID:    tenantID,
		Number:      number,
		Date:        d,
		ReceiptType: inp.ReceiptType,
		EntityID:    entityID,
		Total:       pgNumeric(fmt.Sprintf("%.2f", allocTotal)),
		UserID:      userID,
		Notes:       inp.Notes,
	})
	if err != nil {
		return nil, fmt.Errorf("create receipt: %w", err)
	}

	// 2. Create treasury movements for each payment
	for i, pay := range inp.Payments {
		movType := "cash_in"
		if inp.ReceiptType == "payment" {
			movType = "cash_out"
		}
		switch pay.PaymentMethod {
		case "transfer":
			movType = "bank_deposit"
			if inp.ReceiptType == "payment" {
				movType = "bank_withdrawal"
			}
		case "check":
			movType = "check_received"
			if inp.ReceiptType == "payment" {
				movType = "check_issued"
			}
		}

		mov, err := qtx.CreateTreasuryMovement(ctx, repository.CreateTreasuryMovementParams{
			TenantID:       tenantID,
			Date:           d,
			Number:         fmt.Sprintf("%s-P%d", number, i+1),
			MovementType:   movType,
			Amount:         pgNumeric(pay.Amount),
			BankAccountID:  optUUIDFromPtr(pay.BankAccountID),
			EntityID:       entityID,
			PaymentMethod:  pgText(pay.PaymentMethod),
			ReferenceType:  pgText("receipt"),
			ReferenceID:    receipt.ID,
			UserID:         userID,
			Notes:          pay.Notes,
		})
		if err != nil {
			return nil, fmt.Errorf("create treasury movement %d: %w", i, err)
		}

		_, err = qtx.CreateReceiptPayment(ctx, repository.CreateReceiptPaymentParams{
			TenantID:           tenantID,
			ReceiptID:          receipt.ID,
			PaymentMethod:      pay.PaymentMethod,
			Amount:             pgNumeric(pay.Amount),
			TreasuryMovementID: mov.ID,
			CheckID:            optUUIDFromPtr(pay.CheckID),
			BankAccountID:      optUUIDFromPtr(pay.BankAccountID),
			Notes:              pay.Notes,
		})
		if err != nil {
			return nil, fmt.Errorf("create receipt payment %d: %w", i, err)
		}
	}

	// 3. Create allocations (impute payments to invoices in current accounts)
	direction := "receivable"
	if inp.ReceiptType == "payment" {
		direction = "payable"
	}
	for i, alloc := range inp.Allocations {
		invID, err := parseUUIDStr(alloc.InvoiceID)
		if err != nil {
			return nil, fmt.Errorf("invalid invoice_id in allocation %d: %w", i, err)
		}

		// Create account movement for the payment
		acctMov, err := qtx.CreateAccountMovement(ctx, repository.CreateAccountMovementParams{
			TenantID:    tenantID,
			EntityID:    entityID,
			Date:        d,
			MovementType: "payment",
			Direction:   direction,
			Amount:      negateNumeric(pgNumeric(alloc.Amount)), // payment reduces balance
			Balance:     zeroNumeric(),
			InvoiceID:   invID,
			Notes:       fmt.Sprintf("Recibo %s", number),
			UserID:      userID,
		})
		if err != nil {
			return nil, fmt.Errorf("create account movement %d: %w", i, err)
		}

		_, err = qtx.CreateReceiptAllocation(ctx, repository.CreateReceiptAllocationParams{
			TenantID:          tenantID,
			ReceiptID:         receipt.ID,
			InvoiceID:         invID,
			Amount:            pgNumeric(alloc.Amount),
			AccountMovementID: acctMov.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("create allocation %d: %w", i, err)
		}

		// Reduce the ORIGINAL invoice movement's outstanding balance
		// (not the payment movement we just created — that starts at 0)
		invoiceMovs, err := qtx.ListAccountMovementsByInvoice(ctx, repository.ListAccountMovementsByInvoiceParams{
			TenantID: tenantID, InvoiceID: invID,
		})
		if err != nil {
			return nil, fmt.Errorf("list invoice movements %d: %w", i, err)
		}
		// Find the original invoice movement (movement_type='invoice', has balance > 0)
		for _, im := range invoiceMovs {
			if im.MovementType == "invoice" {
				rows, err := qtx.UpdateMovementBalance(ctx, repository.UpdateMovementBalanceParams{
					ID: im.ID, TenantID: tenantID,
					Balance: pgNumeric(alloc.Amount),
				})
				if err != nil {
					return nil, fmt.Errorf("update invoice balance %d: %w", i, err)
				}
				if rows == 0 {
					return nil, fmt.Errorf("allocation %d exceeds invoice outstanding balance", i)
				}
				break
			}
		}
	}

	// 4. Create withholdings (inline, within same TX)
	for i, wh := range inp.Withholdings {
		whEntityID, _ := parseUUIDStr(wh.EntityID)
		var certNum pgtype.Text
		if wh.CertificateNum != "" {
			certNum = pgText(wh.CertificateNum)
		}
		var whDate pgtype.Date
		_ = whDate.Scan(wh.Date)

		withholding, err := qtx.CreateWithholding(ctx, repository.CreateWithholdingParams{
			TenantID:       tenantID,
			EntityID:       whEntityID,
			Type:           wh.Type,
			Rate:           pgNumeric(wh.Rate),
			BaseAmount:     pgNumeric(wh.BaseAmount),
			Amount:         pgNumeric(wh.Amount),
			CertificateNum: certNum,
			Date:           whDate,
		})
		if err != nil {
			return nil, fmt.Errorf("create withholding %d: %w", i, err)
		}

		_, err = qtx.CreateReceiptWithholding(ctx, repository.CreateReceiptWithholdingParams{
			TenantID:      tenantID,
			ReceiptID:     receipt.ID,
			WithholdingID: withholding.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("create receipt withholding %d: %w", i, err)
		}
	}

	// 5. Generate auto journal entry (double-entry bookkeeping)
	// Collection: debit cash/bank, credit receivable
	// Payment: debit payable, credit cash/bank
	var journalLines []repository.CreateJournalLineParams
	// Debit side: cash/bank accounts from payments
	for i, pay := range inp.Payments {
		journalLines = append(journalLines, repository.CreateJournalLineParams{
			TenantID:  tenantID,
			EntryDate: d,
			Debit:     pgNumeric(pay.Amount),
			Credit:    zeroNumeric(),
			Description: fmt.Sprintf("Recibo %s - %s #%d", number, pay.PaymentMethod, i+1),
			SortOrder: int32(i),
		})
	}
	// Credit side: entity account (receivable/payable)
	journalLines = append(journalLines, repository.CreateJournalLineParams{
		TenantID:  tenantID,
		EntryDate: d,
		Debit:     zeroNumeric(),
		Credit:    pgNumeric(fmt.Sprintf("%.2f", allocTotal)),
		Description: fmt.Sprintf("Recibo %s - imputación", number),
		SortOrder: int32(len(inp.Payments)),
	})
	// Swap debit/credit for payments (we pay supplier, not receive from customer)
	if inp.ReceiptType == "payment" {
		for i := range journalLines {
			journalLines[i].Debit, journalLines[i].Credit = journalLines[i].Credit, journalLines[i].Debit
		}
	}

	// Resolve the open fiscal year for the receipt date — required by the
	// erp_journal_entries.fiscal_year_id NOT NULL constraint.
	fy, err := qtx.GetFiscalYearByDate(ctx, repository.GetFiscalYearByDateParams{
		TenantID: tenantID, Date: d,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve fiscal year for receipt date: %w", err)
	}

	// Create journal entry + lines
	entry, err := qtx.CreateJournalEntry(ctx, repository.CreateJournalEntryParams{
		TenantID:      tenantID,
		Number:        "AUT-" + number,
		Date:          d,
		FiscalYearID:  fy.ID,
		Concept:       fmt.Sprintf("Recibo %s", number),
		EntryType:     "auto",
		ReferenceType: pgText("receipt"),
		ReferenceID:   receipt.ID,
		UserID:        userID,
	})
	if err != nil {
		return nil, fmt.Errorf("create journal entry: %w", err)
	}
	for _, jl := range journalLines {
		jl.EntryID = entry.ID
		if _, err := qtx.CreateJournalLine(ctx, jl); err != nil {
			return nil, fmt.Errorf("create journal line: %w", err)
		}
	}
	// Post the auto entry
	if _, err := qtx.PostJournalEntry(ctx, repository.PostJournalEntryParams{
		ID: entry.ID, TenantID: tenantID,
	}); err != nil {
		return nil, fmt.Errorf("post journal entry: %w", err)
	}
	if _, err := qtx.SetReceiptJournalEntry(ctx, repository.SetReceiptJournalEntryParams{
		ID: receipt.ID, TenantID: tenantID, JournalEntryID: entry.ID,
	}); err != nil {
		return nil, fmt.Errorf("set receipt journal entry: %w", err)
	}

	// StrictLogger
	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.receipt.created", Resource: uuidStr(receipt.ID),
		Details: map[string]any{
			"type":         inp.ReceiptType,
			"number":       number,
			"total":        allocTotal,
			"payments":     len(inp.Payments),
			"allocations":  len(inp.Allocations),
			"withholdings": len(inp.Withholdings),
		}, IP: ip,
	}); err != nil {
		return nil, fmt.Errorf("strict audit failed, aborting: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit receipt: %w", err)
	}

	s.publisher.Broadcast(tenantID, "erp_treasury", map[string]any{
		"action": "receipt_created", "receipt_id": uuidStr(receipt.ID),
	})

	// Fetch detail for response
	detail, _ := s.GetReceipt(ctx, tenantID, receipt.ID)
	return detail, nil
}

// VoidReceipt cancels a confirmed receipt with cascade reversal:
// reverses treasury movements, account movements, and journal entry.
func (s *Treasury) VoidReceipt(ctx context.Context, tenantID string, receiptID pgtype.UUID, userID, ip string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := s.repo.WithTx(tx)

	receipt, err := qtx.GetReceipt(ctx, repository.GetReceiptParams{ID: receiptID, TenantID: tenantID})
	if err != nil {
		return fmt.Errorf("get receipt: %w", err)
	}
	if receipt.Status != "confirmed" {
		return fmt.Errorf("receipt is not confirmed (status: %s)", receipt.Status)
	}

	// 1. Reverse account movements (from allocations)
	allocs, _ := qtx.ListReceiptAllocations(ctx, repository.ListReceiptAllocationsParams{
		TenantID: tenantID, ReceiptID: receiptID,
	})
	for _, alloc := range allocs {
		// Restore original invoice balance
		invoiceMovs, _ := qtx.ListAccountMovementsByInvoice(ctx, repository.ListAccountMovementsByInvoiceParams{
			TenantID: tenantID, InvoiceID: alloc.InvoiceID,
		})
		for _, im := range invoiceMovs {
			if im.MovementType == "invoice" {
				// Restore balance by negating the reduction (passes WHERE balance >= check)
				if _, err := qtx.UpdateMovementBalance(ctx, repository.UpdateMovementBalanceParams{
					ID: im.ID, TenantID: tenantID,
					Balance: negateNumeric(alloc.Amount),
				}); err != nil {
					return fmt.Errorf("restore invoice balance: %w", err)
				}
				break
			}
		}
		// Create reversal account movement
		if _, err := qtx.CreateAccountMovement(ctx, repository.CreateAccountMovementParams{
			TenantID:     tenantID,
			EntityID:     receipt.EntityID,
			Date:         receipt.Date,
			MovementType: "reversal",
			Direction:    "receivable",
			Amount:       alloc.Amount,
			Balance:      zeroNumeric(),
			InvoiceID:    alloc.InvoiceID,
			Notes:        fmt.Sprintf("Reversa recibo %s", receipt.Number),
			UserID:       userID,
		}); err != nil {
			return fmt.Errorf("create reversal movement: %w", err)
		}
	}

	// 2. Reverse journal entry (if exists)
	if receipt.JournalEntryID.Valid {
		// Create reversal entry with swapped debit/credit
		original, err := qtx.GetJournalEntry(ctx, repository.GetJournalEntryParams{
			ID: receipt.JournalEntryID, TenantID: tenantID,
		})
		if err == nil {
			lines, _ := qtx.ListJournalLines(ctx, repository.ListJournalLinesParams{
				EntryID: receipt.JournalEntryID, TenantID: tenantID,
			})
			revEntry, err := qtx.CreateJournalEntry(ctx, repository.CreateJournalEntryParams{
				TenantID:      tenantID,
				Number:        "REV-" + original.Number,
				Date:          receipt.Date,
				FiscalYearID:  original.FiscalYearID,
				Concept:       "Reversa: " + original.Concept,
				EntryType:     "reversal",
				ReferenceType: pgText("reversal"),
				ReferenceID:   receipt.JournalEntryID,
				UserID:        userID,
			})
			if err == nil {
				for i, l := range lines {
					if _, err := qtx.CreateJournalLine(ctx, repository.CreateJournalLineParams{
						TenantID:     tenantID,
						EntryID:      revEntry.ID,
						AccountID:    l.AccountID,
						CostCenterID: l.CostCenterID,
						EntryDate:    receipt.Date,
						Debit:        l.Credit,
						Credit:       l.Debit,
						Description:  "Reversa: " + l.Description,
						SortOrder:    int32(i),
					}); err != nil {
						return fmt.Errorf("create reversal journal line: %w", err)
					}
				}
				if _, err := qtx.PostJournalEntry(ctx, repository.PostJournalEntryParams{ID: revEntry.ID, TenantID: tenantID}); err != nil {
					return fmt.Errorf("post reversal journal entry: %w", err)
				}
				if _, err := qtx.MarkEntryReversed(ctx, repository.MarkEntryReversedParams{
					ID: receipt.JournalEntryID, TenantID: tenantID, ReversedBy: revEntry.ID,
				}); err != nil {
					return fmt.Errorf("mark entry reversed: %w", err)
				}
			}
		}
	}

	// 3. Mark receipt as cancelled
	rows, err := qtx.VoidReceipt(ctx, repository.VoidReceiptParams{
		ID: receiptID, TenantID: tenantID,
	})
	if err != nil {
		return fmt.Errorf("void receipt: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("receipt void failed")
	}

	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.receipt.voided", Resource: uuidStr(receiptID), IP: ip,
	}); err != nil {
		return fmt.Errorf("strict audit failed, aborting: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit void: %w", err)
	}

	s.publisher.Broadcast(tenantID, "erp_treasury", map[string]any{
		"action": "receipt_voided", "receipt_id": uuidStr(receiptID),
	})
	return nil
}

// parseFloatApprox parses a string to float64 for approximate comparisons only
// (e.g., balance pre-check). Never use for financial arithmetic — use pgNumeric.
func parseFloatApprox(s string) float64 {
	if s == "" {
		return 0
	}
	var f float64
	_, _ = fmt.Sscanf(s, "%f", &f)
	return f
}

func optUUIDFromPtr(s *string) pgtype.UUID {
	if s == nil || *s == "" {
		return pgtype.UUID{}
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

// amountsMatch compares a statement line amount against a treasury movement.
// Statement: positive=credit, negative=debit.
// Treasury: movement_type determines direction (deposit=positive, withdrawal=negative).
func amountsMatch(lineAmount, movAmount pgtype.Numeric, movType string) bool {
	la, _ := lineAmount.Float64Value()
	ma, _ := movAmount.Float64Value()

	// Normalize: deposits are positive in both, withdrawals are negative in statement
	var movSigned float64
	switch movType {
	case "bank_deposit", "check_received", "cash_in":
		movSigned = ma.Float64
	case "bank_withdrawal", "check_issued", "cash_out":
		movSigned = -ma.Float64
	default:
		movSigned = ma.Float64
	}

	// Compare with small tolerance for rounding
	diff := la.Float64 - movSigned
	if diff < 0 {
		diff = -diff
	}
	return diff < 0.01
}

// datesClose checks if two dates are within N days of each other.
func datesClose(a, b pgtype.Date, days int) bool {
	if !a.Valid || !b.Valid {
		return false
	}
	diff := a.Time.Sub(b.Time)
	if diff < 0 {
		diff = -diff
	}
	return diff.Hours()/24 <= float64(days)
}

// ============================================================
// Bank imports (bcs_importacion parity)
// ============================================================

func (s *Treasury) ListBankImports(ctx context.Context, tenantID string, accountFilter, processedFilter int32, dateFrom, dateTo pgtype.Date, limit, offset int) ([]repository.ErpBankImport, error) {
	return s.repo.ListBankImports(ctx, repository.ListBankImportsParams{
		TenantID:        tenantID,
		Limit:           int32(limit),
		Offset:          int32(offset),
		AccountFilter:   accountFilter,
		ProcessedFilter: processedFilter,
		DateFrom:        dateFrom,
		DateTo:          dateTo,
	})
}

// UpdateBankImportRequest is the input for UpdateBankImportProcessed.
type UpdateBankImportRequest struct {
	ID                 pgtype.UUID
	TenantID           string
	Processed          int32
	TreasuryMovementID pgtype.UUID
	UserID             string
	IP                 string
}

// UpdateBankImportProcessed toggles the processed flag on a bank-import
// staging row and optionally links / unlinks a treasury movement.
// Parity: bancos_local/bcsmovim_importacion_auto_mov_ins.xml.
func (s *Treasury) UpdateBankImportProcessed(ctx context.Context, req UpdateBankImportRequest) error {
	rows, err := s.repo.UpdateBankImportProcessed(ctx, repository.UpdateBankImportProcessedParams{
		ID:                 req.ID,
		TenantID:           req.TenantID,
		Processed:          req.Processed,
		TreasuryMovementID: req.TreasuryMovementID,
	})
	if err != nil {
		return fmt.Errorf("update bank import: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("bank import not found")
	}
	idStr := uuidStr(req.ID)
	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID: req.TenantID, UserID: req.UserID,
		Action:   "erp.bank_imports.processed_changed",
		Resource: idStr,
		Details: map[string]any{
			"processed":            req.Processed,
			"treasury_movement_id": uuidStr(req.TreasuryMovementID),
		},
		IP: req.IP,
	}); err != nil {
		return fmt.Errorf("strict audit failed, aborting: %w", err)
	}
	s.publisher.Broadcast(req.TenantID, "erp_bank_imports", map[string]any{
		"action":    "processed_changed",
		"id":        idStr,
		"processed": req.Processed,
	})
	return nil
}
