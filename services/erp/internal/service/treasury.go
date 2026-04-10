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

func (s *Treasury) CreateMovement(ctx context.Context, req CreateTreasuryMovementRequest) (repository.ErpTreasuryMovement, error) {
	if !validMovementTypesT[req.MovementType] {
		return repository.ErpTreasuryMovement{}, fmt.Errorf("invalid movement type: %s", req.MovementType)
	}

	mov, err := s.repo.CreateTreasuryMovement(ctx, req.CreateTreasuryMovementParams)
	if err != nil {
		return repository.ErpTreasuryMovement{}, fmt.Errorf("create movement: %w", err)
	}

	idStr := uuidStr(mov.ID)
	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID: req.TenantID, UserID: req.UserIDVal,
		Action: "erp.treasury.movement", Resource: idStr,
		Details: map[string]any{"type": req.MovementType}, IP: req.IP,
	}); err != nil {
		slog.Error("STRICT audit failed on treasury movement", "error", err)
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

func (s *Treasury) UpdateCheckStatus(ctx context.Context, id pgtype.UUID, tenantID, newStatus, userID, ip string) error {
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
