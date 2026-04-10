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
func (s *Accounting) ListFiscalYears(ctx context.Context, tenantID string) ([]repository.ErpFiscalYear, error) {
	return s.repo.ListFiscalYears(ctx, tenantID)
}

// CreateFiscalYear creates a new fiscal year.
func (s *Accounting) CreateFiscalYear(ctx context.Context, tenantID string, year int, startDate, endDate, userID, ip string) (repository.ErpFiscalYear, error) {
	var sd, ed pgtype.Date
	_ = sd.Scan(startDate)
	_ = ed.Scan(endDate)
	fy, err := s.repo.CreateFiscalYear(ctx, repository.CreateFiscalYearParams{
		TenantID: tenantID, Year: int32(year), StartDate: sd, EndDate: ed,
	})
	if err != nil {
		return repository.ErpFiscalYear{}, fmt.Errorf("create fiscal year: %w", err)
	}
	s.auditLog.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.fiscal_year.created", Resource: uuidStr(fy.ID),
		Details: map[string]any{"year": year}, IP: ip,
	})
	return fy, nil
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
