-- name: ListBankAccounts :many
SELECT id, tenant_id, bank_name, branch, account_number, cbu, alias,
       currency_id, account_id, active, created_at
FROM erp_bank_accounts
WHERE tenant_id = $1 AND (sqlc.arg(active_only)::BOOLEAN = false OR active = true)
ORDER BY bank_name;

-- name: CreateBankAccount :one
INSERT INTO erp_bank_accounts (tenant_id, bank_name, branch, account_number, cbu, alias, currency_id, account_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, tenant_id, bank_name, branch, account_number, cbu, alias, currency_id, account_id, active, created_at;

-- name: ListCashRegisters :many
SELECT id, tenant_id, name, account_id, active, created_at
FROM erp_cash_registers WHERE tenant_id = $1 ORDER BY name;

-- name: CreateCashRegister :one
INSERT INTO erp_cash_registers (tenant_id, name, account_id)
VALUES ($1, $2, $3)
RETURNING id, tenant_id, name, account_id, active, created_at;

-- name: ListTreasuryMovements :many
SELECT tm.id, tm.tenant_id, tm.date, tm.number, tm.movement_type, tm.amount,
       tm.currency_id, tm.bank_account_id, tm.cash_register_id, tm.entity_id,
       tm.concept_id, tm.payment_method, tm.user_id, tm.notes, tm.status,
       tm.reconciled, tm.reconciliation_id, tm.created_at,
       e.name AS entity_name
FROM erp_treasury_movements tm
LEFT JOIN erp_entities e ON e.id = tm.entity_id
WHERE tm.tenant_id = $1
  AND (sqlc.arg(date_from)::DATE IS NULL OR tm.date >= sqlc.arg(date_from)::DATE)
  AND (sqlc.arg(date_to)::DATE IS NULL OR tm.date <= sqlc.arg(date_to)::DATE)
  AND (sqlc.arg(type_filter)::TEXT = '' OR tm.movement_type = sqlc.arg(type_filter)::TEXT)
ORDER BY tm.date DESC, tm.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateTreasuryMovement :one
INSERT INTO erp_treasury_movements (tenant_id, date, number, movement_type, amount,
    currency_id, bank_account_id, cash_register_id, entity_id, concept_id,
    payment_method, reference_type, reference_id, journal_entry_id, user_id, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING id, tenant_id, date, number, movement_type, amount, currency_id,
    bank_account_id, cash_register_id, entity_id, concept_id, payment_method,
    reference_type, reference_id, journal_entry_id, user_id, notes, status, created_at;

-- name: ListChecks :many
SELECT id, tenant_id, direction, number, bank_name, amount, issue_date, due_date,
       entity_id, status, movement_id, notes, created_at
FROM erp_checks
WHERE tenant_id = $1
  AND (sqlc.arg(direction_filter)::TEXT = '' OR direction = sqlc.arg(direction_filter)::TEXT)
  AND (sqlc.arg(status_filter)::TEXT = '' OR status = sqlc.arg(status_filter)::TEXT)
ORDER BY due_date;

-- name: CreateCheck :one
INSERT INTO erp_checks (tenant_id, direction, number, bank_name, amount, issue_date, due_date, entity_id, movement_id, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, tenant_id, direction, number, bank_name, amount, issue_date, due_date, entity_id, status, movement_id, notes, created_at;

-- name: GetCheck :one
SELECT id, tenant_id, direction, number, bank_name, amount, issue_date, due_date,
       entity_id, status, movement_id, notes, created_at
FROM erp_checks WHERE id = $1 AND tenant_id = $2;

-- name: UpdateCheckStatus :execrows
UPDATE erp_checks SET status = $3
WHERE id = $1 AND tenant_id = $2;

-- name: ListCashCounts :many
SELECT id, tenant_id, cash_register_id, date, expected, counted, difference, user_id, notes, created_at
FROM erp_cash_counts WHERE tenant_id = $1 ORDER BY date DESC LIMIT $2 OFFSET $3;

-- name: CreateCashCount :one
INSERT INTO erp_cash_counts (tenant_id, cash_register_id, date, expected, counted, difference, user_id, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, tenant_id, cash_register_id, date, expected, counted, difference, user_id, notes, created_at;

-- name: GetTreasuryBalance :many
SELECT ba.id, ba.bank_name, ba.account_number,
       COALESCE(SUM(CASE WHEN tm.movement_type IN ('bank_deposit', 'check_received') THEN tm.amount ELSE 0 END), 0)::NUMERIC(16,2) AS total_in,
       COALESCE(SUM(CASE WHEN tm.movement_type IN ('bank_withdrawal', 'check_issued') THEN tm.amount ELSE 0 END), 0)::NUMERIC(16,2) AS total_out,
       COALESCE(SUM(CASE WHEN tm.movement_type IN ('bank_deposit', 'check_received') THEN tm.amount
                         WHEN tm.movement_type IN ('bank_withdrawal', 'check_issued') THEN -tm.amount
                         ELSE 0 END), 0)::NUMERIC(16,2) AS balance
FROM erp_bank_accounts ba
LEFT JOIN erp_treasury_movements tm ON tm.bank_account_id = ba.id AND tm.status = 'confirmed'
WHERE ba.tenant_id = $1 AND ba.active = true
GROUP BY ba.id, ba.bank_name, ba.account_number
ORDER BY ba.bank_name;

-- ============================================================
-- Reconciliation queries (Plan 18 Fase 1)
-- ============================================================

-- name: CreateReconciliation :one
INSERT INTO erp_bank_reconciliations (tenant_id, bank_account_id, period, statement_balance, book_balance, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tenant_id, bank_account_id, period, statement_balance, book_balance, status, user_id, confirmed_at, created_at;

-- name: GetReconciliation :one
SELECT r.id, r.tenant_id, r.bank_account_id, r.period, r.statement_balance,
       r.book_balance, r.status, r.user_id, r.confirmed_at, r.created_at,
       ba.bank_name, ba.account_number
FROM erp_bank_reconciliations r
JOIN erp_bank_accounts ba ON ba.id = r.bank_account_id
WHERE r.id = $1 AND r.tenant_id = $2;

-- name: ListReconciliations :many
SELECT r.id, r.tenant_id, r.bank_account_id, r.period, r.statement_balance,
       r.book_balance, r.status, r.confirmed_at, r.created_at,
       ba.bank_name, ba.account_number
FROM erp_bank_reconciliations r
JOIN erp_bank_accounts ba ON ba.id = r.bank_account_id
WHERE r.tenant_id = $1
ORDER BY r.period DESC, ba.bank_name;

-- name: CreateStatementLine :one
INSERT INTO erp_bank_statement_lines (tenant_id, reconciliation_id, date, description, amount, reference)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tenant_id, reconciliation_id, date, description, amount, reference, matched, movement_id, created_at;

-- name: ListStatementLines :many
SELECT id, tenant_id, reconciliation_id, date, description, amount, reference, matched, movement_id, created_at
FROM erp_bank_statement_lines
WHERE tenant_id = $1 AND reconciliation_id = $2
ORDER BY date, created_at;

-- name: ListUnmatchedStatementLines :many
SELECT id, tenant_id, reconciliation_id, date, description, amount, reference
FROM erp_bank_statement_lines
WHERE tenant_id = $1 AND reconciliation_id = $2 AND matched = false
ORDER BY date;

-- name: ListUnreconciledMovements :many
SELECT tm.id, tm.date, tm.number, tm.movement_type, tm.amount, tm.notes
FROM erp_treasury_movements tm
WHERE tm.tenant_id = $1 AND tm.bank_account_id = $2 AND tm.reconciled = false
  AND tm.status = 'confirmed'
  AND to_char(tm.date, 'YYYY-MM') = sqlc.arg(period)::TEXT
ORDER BY tm.date;

-- name: MatchStatementLine :execrows
UPDATE erp_bank_statement_lines SET matched = true, movement_id = $3
WHERE id = $1 AND tenant_id = $2 AND matched = false;

-- name: MarkMovementReconciled :execrows
UPDATE erp_treasury_movements SET reconciled = true, reconciliation_id = $3
WHERE id = $1 AND tenant_id = $2 AND reconciled = false;

-- name: ConfirmReconciliation :execrows
UPDATE erp_bank_reconciliations SET status = 'confirmed', confirmed_at = now()
WHERE id = $1 AND tenant_id = $2 AND status = 'draft';
