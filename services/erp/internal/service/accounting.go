package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

// Accounting handles accounting business logic.
type Accounting struct {
	repo      *repository.Queries
	pool      TxStarter
	audit     audit.StrictLogger
	auditLog  *audit.Writer
	publisher *traces.Publisher
}

// TxStarter starts a database transaction. Implemented by pgxpool.Pool.
type TxStarter interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// NewAccounting creates an accounting service.
// Uses StrictLogger for financial operations (fail-closed).
func NewAccounting(repo *repository.Queries, pool TxStarter, auditWriter *audit.Writer, publisher *traces.Publisher) *Accounting {
	return &Accounting{
		repo: repo, pool: pool, audit: auditWriter, auditLog: auditWriter, publisher: publisher,
	}
}

// ListAccounts returns the chart of accounts.
func (s *Accounting) ListAccounts(ctx context.Context, tenantID string, activeOnly bool) ([]repository.ErpAccount, error) {
	return s.repo.ListAccounts(ctx, repository.ListAccountsParams{
		TenantID: tenantID, ActiveOnly: activeOnly,
	})
}

// CreateAccount creates a new account in the chart.
func (s *Accounting) CreateAccount(ctx context.Context, tenantID, code, name string, parentID pgtype.UUID, accountType string, isDetail bool, costCenterID pgtype.UUID, userID, ip string) (repository.ErpAccount, error) {
	acct, err := s.repo.CreateAccount(ctx, repository.CreateAccountParams{
		TenantID: tenantID, Code: code, Name: name, ParentID: parentID,
		AccountType: accountType, IsDetail: isDetail, CostCenterID: costCenterID,
	})
	if err != nil {
		return repository.ErpAccount{}, fmt.Errorf("create account: %w", err)
	}
	s.auditLog.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.account.created", Resource: uuidStr(acct.ID),
		Details: map[string]any{"code": code, "type": accountType}, IP: ip,
	})
	return acct, nil
}

// ListCostCenters returns cost centers.
func (s *Accounting) ListCostCenters(ctx context.Context, tenantID string, activeOnly bool) ([]repository.ErpCostCenter, error) {
	return s.repo.ListCostCenters(ctx, repository.ListCostCentersParams{
		TenantID: tenantID, ActiveOnly: activeOnly,
	})
}

// CreateCostCenter creates a new cost center.
func (s *Accounting) CreateCostCenter(ctx context.Context, tenantID, code, name string, parentID pgtype.UUID, userID, ip string) (repository.ErpCostCenter, error) {
	cc, err := s.repo.CreateCostCenter(ctx, repository.CreateCostCenterParams{
		TenantID: tenantID, Code: code, Name: name, ParentID: parentID,
	})
	if err != nil {
		return repository.ErpCostCenter{}, fmt.Errorf("create cost center: %w", err)
	}
	s.auditLog.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.cost_center.created", Resource: uuidStr(cc.ID), IP: ip,
	})
	return cc, nil
}

// ListFiscalYears returns fiscal years.
func (s *Accounting) ListFiscalYears(ctx context.Context, tenantID string) ([]repository.ListFiscalYearsRow, error) {
	return s.repo.ListFiscalYears(ctx, tenantID)
}

// CreateFiscalYear creates a new fiscal year.
func (s *Accounting) CreateFiscalYear(ctx context.Context, tenantID string, year int, startDate, endDate, userID, ip string) (repository.CreateFiscalYearRow, error) {
	var sd, ed pgtype.Date
	_ = sd.Scan(startDate)
	_ = ed.Scan(endDate)
	fy, err := s.repo.CreateFiscalYear(ctx, repository.CreateFiscalYearParams{
		TenantID: tenantID, Year: int32(year), StartDate: sd, EndDate: ed,
	})
	if err != nil {
		return repository.CreateFiscalYearRow{}, fmt.Errorf("create fiscal year: %w", err)
	}
	s.auditLog.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.fiscal_year.created", Resource: uuidStr(fy.ID),
		Details: map[string]any{"year": year}, IP: ip,
	})
	return fy, nil
}

// SetFiscalYearResultAccount sets the result account for a fiscal year.
func (s *Accounting) SetFiscalYearResultAccount(ctx context.Context, tenantID string, yearID, accountID pgtype.UUID, userID, ip string) error {
	rows, err := s.repo.SetFiscalYearResultAccount(ctx, repository.SetFiscalYearResultAccountParams{
		ID: yearID, TenantID: tenantID, ResultAccountID: accountID,
	})
	if err != nil {
		return fmt.Errorf("set result account: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("fiscal year not found or not open")
	}
	s.auditLog.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.fiscal_year.result_account_set", Resource: uuidStr(yearID),
		Details: map[string]any{"account_id": uuidStr(accountID)}, IP: ip,
	})
	return nil
}

// CloseResult holds the result of closing a fiscal year.
type CloseResult struct {
	ClosingEntryID pgtype.UUID `json:"closing_entry_id"`
	OpeningEntryID pgtype.UUID `json:"opening_entry_id"`
	NewYearID      pgtype.UUID `json:"new_year_id"`
}

// PreviewClose returns account balances that would be closed.
func (s *Accounting) PreviewClose(ctx context.Context, tenantID string, yearID pgtype.UUID) (*PreviewCloseResult, error) {
	fy, err := s.repo.GetFiscalYear(ctx, repository.GetFiscalYearParams{ID: yearID, TenantID: tenantID})
	if err != nil {
		return nil, fmt.Errorf("get fiscal year: %w", err)
	}
	if fy.Status != "open" {
		return nil, fmt.Errorf("fiscal year is not open")
	}
	if !fy.ResultAccountID.Valid {
		return nil, fmt.Errorf("result_account_id not set — call PATCH /result-account first")
	}

	drafts, err := s.repo.ListDraftEntriesInPeriod(ctx, repository.ListDraftEntriesInPeriodParams{
		TenantID: tenantID, FiscalYearID: yearID,
	})
	if err != nil {
		return nil, fmt.Errorf("list drafts: %w", err)
	}

	balances, err := s.repo.GetAccountBalancesForClose(ctx, repository.GetAccountBalancesForCloseParams{
		TenantID: tenantID, FiscalYearID: yearID,
	})
	if err != nil {
		return nil, fmt.Errorf("get balances: %w", err)
	}

	return &PreviewCloseResult{
		FiscalYear:    fy,
		DraftEntries:  drafts,
		Balances:      balances,
		CanClose:      len(drafts) == 0,
		BlockedReason: previewBlockReason(drafts),
	}, nil
}

// PreviewCloseResult holds the preview data.
type PreviewCloseResult struct {
	FiscalYear    repository.GetFiscalYearRow                `json:"fiscal_year"`
	DraftEntries  []repository.ListDraftEntriesInPeriodRow   `json:"draft_entries"`
	Balances      []repository.GetAccountBalancesForCloseRow `json:"balances"`
	CanClose      bool                                       `json:"can_close"`
	BlockedReason string                                     `json:"blocked_reason,omitempty"`
}

func previewBlockReason(drafts []repository.ListDraftEntriesInPeriodRow) string {
	if len(drafts) == 0 {
		return ""
	}
	return fmt.Sprintf("%d asientos en borrador deben ser contabilizados o eliminados antes del cierre", len(drafts))
}

// CloseFiscalYear closes a fiscal year: generates closing entry (result → equity),
// opening entry for the new year (patrimonial balances), creates the new year.
// Atomic transaction with StrictLogger.
func (s *Accounting) CloseFiscalYear(ctx context.Context, tenantID string, yearID pgtype.UUID, userID, ip string) (*CloseResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := s.repo.WithTx(tx)

	// 1. Validate fiscal year
	fy, err := qtx.GetFiscalYear(ctx, repository.GetFiscalYearParams{ID: yearID, TenantID: tenantID})
	if err != nil {
		return nil, fmt.Errorf("get fiscal year: %w", err)
	}
	if fy.Status != "open" {
		return nil, fmt.Errorf("fiscal year is not open")
	}
	if !fy.ResultAccountID.Valid {
		return nil, fmt.Errorf("result_account_id not set — call PATCH /result-account first")
	}

	// 2. Check no drafts
	drafts, err := qtx.ListDraftEntriesInPeriod(ctx, repository.ListDraftEntriesInPeriodParams{
		TenantID: tenantID, FiscalYearID: yearID,
	})
	if err != nil {
		return nil, fmt.Errorf("list drafts: %w", err)
	}
	if len(drafts) > 0 {
		return nil, fmt.Errorf("%d draft entries must be posted or deleted before closing", len(drafts))
	}

	// 3. Get balances
	balances, err := qtx.GetAccountBalancesForClose(ctx, repository.GetAccountBalancesForCloseParams{
		TenantID: tenantID, FiscalYearID: yearID,
	})
	if err != nil {
		return nil, fmt.Errorf("get balances: %w", err)
	}

	// 4. Generate closing entry: zero out income/expense → result account
	closingEntry, err := s.createClosingEntry(ctx, qtx, tenantID, fy, balances, userID)
	if err != nil {
		return nil, fmt.Errorf("create closing entry: %w", err)
	}

	// 5. Create new fiscal year
	newYear, err := qtx.CreateFiscalYear(ctx, repository.CreateFiscalYearParams{
		TenantID: tenantID, Year: fy.Year + 1,
		StartDate: nextDay(fy.EndDate), EndDate: addYear(fy.EndDate),
		ResultAccountID: fy.ResultAccountID,
	})
	if err != nil {
		return nil, fmt.Errorf("create new fiscal year: %w", err)
	}

	// 6. Generate opening entry: carry patrimonial balances
	openingEntry, err := s.createOpeningEntry(ctx, qtx, tenantID, newYear, balances, userID)
	if err != nil {
		return nil, fmt.Errorf("create opening entry: %w", err)
	}

	// 7. Close the fiscal year
	rows, err := qtx.CloseFiscalYear(ctx, repository.CloseFiscalYearParams{
		ID: yearID, TenantID: tenantID,
		ClosedBy:       pgText(userID),
		ClosingEntryID: closingEntry.ID,
		OpeningEntryID: openingEntry.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("close fiscal year: %w", err)
	}
	if rows == 0 {
		return nil, fmt.Errorf("fiscal year close failed (concurrent modification?)")
	}

	// StrictLogger
	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.fiscal_year.closed", Resource: uuidStr(yearID),
		Details: map[string]any{
			"year":             fy.Year,
			"closing_entry_id": uuidStr(closingEntry.ID),
			"opening_entry_id": uuidStr(openingEntry.ID),
			"new_year_id":      uuidStr(newYear.ID),
		}, IP: ip,
	}); err != nil {
		return nil, fmt.Errorf("strict audit failed, aborting: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit close: %w", err)
	}

	s.publisher.Broadcast(tenantID, "erp_accounting", map[string]any{
		"action": "fiscal_year_closed", "year": fy.Year,
	})

	return &CloseResult{
		ClosingEntryID: closingEntry.ID,
		OpeningEntryID: openingEntry.ID,
		NewYearID:      newYear.ID,
	}, nil
}

// createClosingEntry generates the closing entry that zeros out income/expense accounts
// and moves the net result to the result equity account.
func (s *Accounting) createClosingEntry(ctx context.Context, qtx *repository.Queries, tenantID string, fy repository.GetFiscalYearRow, balances []repository.GetAccountBalancesForCloseRow, userID string) (repository.ErpJournalEntry, error) {
	entry, err := qtx.CreateJournalEntry(ctx, repository.CreateJournalEntryParams{
		TenantID: tenantID,
		Number:   fmt.Sprintf("CIERRE-%d", fy.Year),
		Date:     fy.EndDate,
		FiscalYearID: fy.ID,
		Concept:  fmt.Sprintf("Cierre de ejercicio %d", fy.Year),
		EntryType: "auto",
		UserID:   userID,
	})
	if err != nil {
		return repository.ErpJournalEntry{}, fmt.Errorf("create closing entry header: %w", err)
	}

	// Zero out income and expense accounts
	var netResult pgtype.Numeric
	_ = netResult.Scan("0")
	sortOrder := int32(0)

	for _, b := range balances {
		if b.AccountType != "income" && b.AccountType != "expense" {
			continue
		}
		// Income: normally credit balance (balance < 0) → debit to zero
		// Expense: normally debit balance (balance > 0) → credit to zero
		var debit, credit pgtype.Numeric
		if isPositive(b.Balance) {
			credit = b.Balance
			debit = zeroNumeric()
		} else {
			debit = absNumeric(b.Balance)
			credit = zeroNumeric()
		}

		_, err := qtx.CreateJournalLine(ctx, repository.CreateJournalLineParams{
			TenantID:  tenantID,
			EntryID:   entry.ID,
			AccountID: b.ID,
			EntryDate: fy.EndDate,
			Debit:     debit,
			Credit:    credit,
			Description: fmt.Sprintf("Cierre %s", b.Name),
			SortOrder: sortOrder,
		})
		if err != nil {
			return repository.ErpJournalEntry{}, fmt.Errorf("closing line for %s: %w", b.Code, err)
		}
		netResult = addNumeric(netResult, b.Balance)
		sortOrder++
	}

	// Balancing line to result account (equity)
	var resultDebit, resultCredit pgtype.Numeric
	if isPositive(netResult) {
		// Net income (income > expense): credit result account
		resultDebit = zeroNumeric()
		resultCredit = netResult
	} else {
		// Net loss: debit result account
		resultDebit = absNumeric(netResult)
		resultCredit = zeroNumeric()
	}

	_, err = qtx.CreateJournalLine(ctx, repository.CreateJournalLineParams{
		TenantID:  tenantID,
		EntryID:   entry.ID,
		AccountID: fy.ResultAccountID,
		EntryDate: fy.EndDate,
		Debit:     resultDebit,
		Credit:    resultCredit,
		Description: fmt.Sprintf("Resultado del ejercicio %d", fy.Year),
		SortOrder: sortOrder,
	})
	if err != nil {
		return repository.ErpJournalEntry{}, fmt.Errorf("result line: %w", err)
	}

	// Post the entry immediately (closing entries are always posted)
	_, err = qtx.PostJournalEntry(ctx, repository.PostJournalEntryParams{
		ID: entry.ID, TenantID: tenantID,
	})
	if err != nil {
		return repository.ErpJournalEntry{}, fmt.Errorf("post closing entry: %w", err)
	}

	return entry, nil
}

// createOpeningEntry generates the opening entry for the new fiscal year,
// carrying forward patrimonial (asset, liability, equity) balances.
func (s *Accounting) createOpeningEntry(ctx context.Context, qtx *repository.Queries, tenantID string, newYear repository.CreateFiscalYearRow, balances []repository.GetAccountBalancesForCloseRow, userID string) (repository.ErpJournalEntry, error) {
	entry, err := qtx.CreateJournalEntry(ctx, repository.CreateJournalEntryParams{
		TenantID:     tenantID,
		Number:       fmt.Sprintf("APERTURA-%d", newYear.Year),
		Date:         newYear.StartDate,
		FiscalYearID: newYear.ID,
		Concept:      fmt.Sprintf("Apertura de ejercicio %d", newYear.Year),
		EntryType:    "auto",
		UserID:       userID,
	})
	if err != nil {
		return repository.ErpJournalEntry{}, fmt.Errorf("create opening entry header: %w", err)
	}

	sortOrder := int32(0)
	for _, b := range balances {
		if b.AccountType == "income" || b.AccountType == "expense" {
			continue // skip result accounts — they were zeroed in the closing entry
		}
		// Patrimonial: asset, liability, equity — carry the balance forward
		var debit, credit pgtype.Numeric
		if isPositive(b.Balance) {
			debit = b.Balance
			credit = zeroNumeric()
		} else {
			debit = zeroNumeric()
			credit = absNumeric(b.Balance)
		}

		_, err := qtx.CreateJournalLine(ctx, repository.CreateJournalLineParams{
			TenantID:  tenantID,
			EntryID:   entry.ID,
			AccountID: b.ID,
			EntryDate: newYear.StartDate,
			Debit:     debit,
			Credit:    credit,
			Description: fmt.Sprintf("Apertura %s", b.Name),
			SortOrder: sortOrder,
		})
		if err != nil {
			return repository.ErpJournalEntry{}, fmt.Errorf("opening line for %s: %w", b.Code, err)
		}
		sortOrder++
	}

	// Post immediately
	_, err = qtx.PostJournalEntry(ctx, repository.PostJournalEntryParams{
		ID: entry.ID, TenantID: tenantID,
	})
	if err != nil {
		return repository.ErpJournalEntry{}, fmt.Errorf("post opening entry: %w", err)
	}

	return entry, nil
}

// Numeric helpers for fiscal year close operations.
func zeroNumeric() pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan("0")
	return n
}

func absNumeric(n pgtype.Numeric) pgtype.Numeric {
	// pgtype.Numeric stores value internally — negate if negative
	f, _ := n.Float64Value()
	if f.Float64 < 0 {
		var result pgtype.Numeric
		_ = result.Scan(fmt.Sprintf("%f", -f.Float64))
		return result
	}
	return n
}

func isPositive(n pgtype.Numeric) bool {
	f, _ := n.Float64Value()
	return f.Float64 > 0
}

func addNumeric(a, b pgtype.Numeric) pgtype.Numeric {
	fa, _ := a.Float64Value()
	fb, _ := b.Float64Value()
	var result pgtype.Numeric
	_ = result.Scan(fmt.Sprintf("%f", fa.Float64+fb.Float64))
	return result
}

func pgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func nextDay(d pgtype.Date) pgtype.Date {
	if !d.Valid {
		return d
	}
	next := d.Time.AddDate(0, 0, 1)
	return pgtype.Date{Time: next, Valid: true}
}

func addYear(d pgtype.Date) pgtype.Date {
	if !d.Valid {
		return d
	}
	next := d.Time.AddDate(1, 0, 0)
	return pgtype.Date{Time: next, Valid: true}
}

// EntryDetail bundles a journal entry with its lines.
type EntryDetail struct {
	Entry repository.ErpJournalEntry   `json:"entry"`
	Lines []repository.ListJournalLinesRow `json:"lines"`
}

// ListEntries returns paginated journal entries.
func (s *Accounting) ListEntries(ctx context.Context, tenantID string, dateFrom, dateTo pgtype.Date, status string, limit, offset int) ([]repository.ErpJournalEntry, error) {
	return s.repo.ListJournalEntries(ctx, repository.ListJournalEntriesParams{
		TenantID: tenantID, DateFrom: dateFrom, DateTo: dateTo,
		StatusFilter: status, Limit: int32(limit), Offset: int32(offset),
	})
}

// GetEntry returns a journal entry with lines.
func (s *Accounting) GetEntry(ctx context.Context, id pgtype.UUID, tenantID string) (*EntryDetail, error) {
	entry, err := s.repo.GetJournalEntry(ctx, repository.GetJournalEntryParams{ID: id, TenantID: tenantID})
	if err != nil {
		return nil, fmt.Errorf("get entry: %w", err)
	}
	lines, err := s.repo.ListJournalLines(ctx, repository.ListJournalLinesParams{EntryID: id, TenantID: tenantID})
	if err != nil {
		return nil, fmt.Errorf("list lines: %w", err)
	}
	return &EntryDetail{Entry: entry, Lines: lines}, nil
}

// CreateEntryRequest holds data for creating a journal entry with lines.
type CreateEntryRequest struct {
	TenantID     string
	Number       string
	Date         pgtype.Date
	FiscalYearID pgtype.UUID
	Concept      string
	EntryType    string
	RefType      pgtype.Text
	RefID        pgtype.UUID
	UserID       string
	IP           string
	Lines        []CreateLineRequest
}

// CreateLineRequest holds data for one journal line.
type CreateLineRequest struct {
	AccountID    pgtype.UUID
	CostCenterID pgtype.UUID
	Debit        string
	Credit       string
	Description  string
}

var validEntryTypes = map[string]bool{"manual": true, "auto": true, "adjustment": true}
var validAccountTypes = map[string]bool{"asset": true, "liability": true, "equity": true, "income": true, "expense": true}

// CreateEntry creates a journal entry with lines in a single transaction.
// Uses StrictLogger (pattern P7) — if audit fails, transaction is rolled back.
func (s *Accounting) CreateEntry(ctx context.Context, req CreateEntryRequest) (*EntryDetail, error) {
	if req.Concept == "" || len(req.Lines) == 0 {
		return nil, fmt.Errorf("concept and at least one line required")
	}
	if req.EntryType == "" {
		req.EntryType = "manual"
	}
	if !validEntryTypes[req.EntryType] {
		return nil, fmt.Errorf("invalid entry type: %s", req.EntryType)
	}

	// Transaction: entry + all lines atomically
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.repo.WithTx(tx)

	entry, err := qtx.CreateJournalEntry(ctx, repository.CreateJournalEntryParams{
		TenantID: req.TenantID, Number: req.Number, Date: req.Date,
		FiscalYearID: req.FiscalYearID, Concept: req.Concept,
		EntryType: req.EntryType, ReferenceType: req.RefType,
		ReferenceID: req.RefID, UserID: req.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("create entry: %w", err)
	}

	for i, l := range req.Lines {
		_, err := qtx.CreateJournalLine(ctx, repository.CreateJournalLineParams{
			TenantID:     req.TenantID,
			EntryID:      entry.ID,
			AccountID:    l.AccountID,
			CostCenterID: l.CostCenterID,
			EntryDate:    req.Date,
			Debit:        pgNumeric(l.Debit),
			Credit:       pgNumeric(l.Credit),
			Description:  l.Description,
			SortOrder:    int32(i),
		})
		if err != nil {
			return nil, fmt.Errorf("create line %d: %w", i, err)
		}
	}

	// StrictLogger — if audit fails, abort transaction (pattern P7)
	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID: req.TenantID, UserID: req.UserID,
		Action: "erp.journal.created", Resource: uuidStr(entry.ID),
		Details: map[string]any{"number": req.Number, "lines": len(req.Lines)}, IP: req.IP,
	}); err != nil {
		return nil, fmt.Errorf("strict audit failed, aborting: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit entry: %w", err)
	}

	// Fetch lines with JOINed account data for the response
	lines, _ := s.repo.ListJournalLines(ctx, repository.ListJournalLinesParams{
		EntryID: entry.ID, TenantID: req.TenantID,
	})

	s.publisher.Broadcast(req.TenantID, "erp_accounting", map[string]any{
		"action": "entry_created", "entry_id": uuidStr(entry.ID),
	})

	return &EntryDetail{Entry: repository.ErpJournalEntry(entry), Lines: lines}, nil
}

// PostEntry posts a draft journal entry (immutable after posting).
func (s *Accounting) PostEntry(ctx context.Context, id pgtype.UUID, tenantID, userID, ip string) error {
	rows, err := s.repo.PostJournalEntry(ctx, repository.PostJournalEntryParams{
		ID: id, TenantID: tenantID,
	})
	if err != nil {
		return fmt.Errorf("post entry: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("entry not found or already posted")
	}

	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.journal.posted", Resource: uuidStr(id), IP: ip,
	}); err != nil {
		slog.Error("STRICT audit failed on journal post", "error", err)
	}

	s.publisher.Broadcast(tenantID, "erp_accounting", map[string]any{
		"action": "entry_posted", "entry_id": uuidStr(id),
	})

	return nil
}

// GetBalance returns account balances for a date range.
func (s *Accounting) GetBalance(ctx context.Context, tenantID string, dateFrom, dateTo pgtype.Date) ([]repository.GetAccountBalanceRow, error) {
	return s.repo.GetAccountBalance(ctx, repository.GetAccountBalanceParams{
		TenantID: tenantID, DateFrom: dateFrom, DateTo: dateTo,
	})
}

// GetLedger returns ledger entries for one account.
func (s *Accounting) GetLedger(ctx context.Context, tenantID string, accountID pgtype.UUID, dateFrom, dateTo pgtype.Date, limit, offset int) ([]repository.GetLedgerRow, error) {
	return s.repo.GetLedger(ctx, repository.GetLedgerParams{
		TenantID: tenantID, AccountID: accountID,
		DateFrom: dateFrom, DateTo: dateTo,
		Limit: int32(limit), Offset: int32(offset),
	})
}
