package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

type CurrentAccounts struct {
	repo      *repository.Queries
	pool      TxStarter
	audit     audit.StrictLogger
	auditLog  *audit.Writer
	publisher *traces.Publisher
}

func NewCurrentAccounts(repo *repository.Queries, pool TxStarter, auditWriter *audit.Writer, publisher *traces.Publisher) *CurrentAccounts {
	return &CurrentAccounts{repo: repo, pool: pool, audit: auditWriter, auditLog: auditWriter, publisher: publisher}
}

func (s *CurrentAccounts) ListMovements(ctx context.Context, tenantID string, entityID pgtype.UUID, direction string, dateFrom, dateTo pgtype.Date, limit, offset int) ([]repository.ListAccountMovementsRow, error) {
	return s.repo.ListAccountMovements(ctx, repository.ListAccountMovementsParams{
		TenantID: tenantID, EntityFilter: entityID, DirectionFilter: direction,
		DateFrom: dateFrom, DateTo: dateTo, Limit: int32(limit), Offset: int32(offset),
	})
}

func (s *CurrentAccounts) GetBalances(ctx context.Context, tenantID, direction string) ([]repository.GetEntityBalancesRow, error) {
	return s.repo.GetEntityBalances(ctx, repository.GetEntityBalancesParams{
		TenantID: tenantID, DirectionFilter: direction,
	})
}

func (s *CurrentAccounts) GetOverdue(ctx context.Context, tenantID string) ([]repository.GetOverdueInvoicesRow, error) {
	return s.repo.GetOverdueInvoices(ctx, tenantID)
}

// AllocateRequest holds data for allocating a payment to invoices.
type AllocateRequest struct {
	TenantID  string
	PaymentID pgtype.UUID
	InvoiceID pgtype.UUID
	Amount    string
	UserID    string
	IP        string
}

// Allocate assigns a payment to an invoice, reducing both balances in a transaction.
func (s *CurrentAccounts) Allocate(ctx context.Context, req AllocateRequest) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := s.repo.WithTx(tx)

	amt := pgNumeric(req.Amount)

	_, err = qtx.CreatePaymentAllocation(ctx, repository.CreatePaymentAllocationParams{
		TenantID: req.TenantID, PaymentID: req.PaymentID,
		InvoiceID: req.InvoiceID, Amount: amt,
	})
	if err != nil {
		return fmt.Errorf("create allocation: %w", err)
	}

	// Reduce balance on both payment and invoice movements
	payRows, err := qtx.UpdateMovementBalance(ctx, repository.UpdateMovementBalanceParams{
		ID: req.PaymentID, TenantID: req.TenantID, Balance: amt,
	})
	if err != nil {
		return fmt.Errorf("update payment balance: %w", err)
	}
	if payRows == 0 {
		return fmt.Errorf("insufficient payment balance")
	}

	invRows, err := qtx.UpdateMovementBalance(ctx, repository.UpdateMovementBalanceParams{
		ID: req.InvoiceID, TenantID: req.TenantID, Balance: amt,
	})
	if err != nil {
		return fmt.Errorf("update invoice balance: %w", err)
	}
	if invRows == 0 {
		return fmt.Errorf("insufficient invoice balance")
	}

	// StrictLogger before commit — abort if audit fails (financial operation)
	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID: req.TenantID, UserID: req.UserID,
		Action: "erp.accounts.allocated", Resource: uuidStr(req.PaymentID),
		Details: map[string]any{"invoice_id": uuidStr(req.InvoiceID), "amount": req.Amount}, IP: req.IP,
	}); err != nil {
		return fmt.Errorf("strict audit failed, aborting: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit allocation: %w", err)
	}

	s.publisher.Broadcast(req.TenantID, "erp_accounts", map[string]any{
		"action": "allocated", "payment_id": uuidStr(req.PaymentID),
	})
	return nil
}
