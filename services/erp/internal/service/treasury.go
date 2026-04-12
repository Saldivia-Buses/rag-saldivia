package service

import (
	"context"
	"fmt"

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
	defer tx.Rollback(ctx)
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
	defer tx.Rollback(ctx)
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
	defer tx.Rollback(ctx)
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
				qtx.MatchStatementLine(ctx, repository.MatchStatementLineParams{
					ID: line.ID, TenantID: tenantID, MovementID: mov.ID,
				})
				qtx.MarkMovementReconciled(ctx, repository.MarkMovementReconciledParams{
					ID: mov.ID, TenantID: tenantID, ReconciliationID: reconID,
				})
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
	defer tx.Rollback(ctx)
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

// ConfirmReconciliation confirms a reconciliation.
func (s *Treasury) ConfirmReconciliation(ctx context.Context, tenantID string, reconID pgtype.UUID, userID, ip string) error {
	rows, err := s.repo.ConfirmReconciliation(ctx, repository.ConfirmReconciliationParams{
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
		return fmt.Errorf("strict audit failed: %w", err)
	}

	s.publisher.Broadcast(tenantID, "erp_treasury", map[string]any{
		"action": "reconciliation_confirmed", "reconciliation_id": uuidStr(reconID),
	})
	return nil
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
