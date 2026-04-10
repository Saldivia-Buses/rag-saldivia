-- name: ListAccounts :many
SELECT id, tenant_id, code, name, parent_id, account_type, is_detail, cost_center_id, active
FROM erp_accounts
WHERE tenant_id = $1 AND (sqlc.arg(active_only)::BOOLEAN = false OR active = true)
ORDER BY code;

-- name: GetAccount :one
SELECT id, tenant_id, code, name, parent_id, account_type, is_detail, cost_center_id, active
FROM erp_accounts WHERE id = $1 AND tenant_id = $2;

-- name: CreateAccount :one
INSERT INTO erp_accounts (tenant_id, code, name, parent_id, account_type, is_detail, cost_center_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, code, name, parent_id, account_type, is_detail, cost_center_id, active;

-- name: UpdateAccount :one
UPDATE erp_accounts
SET code = $3, name = $4, parent_id = $5, account_type = $6, is_detail = $7,
    cost_center_id = $8, active = $9
WHERE id = $1 AND tenant_id = $2
RETURNING id, tenant_id, code, name, parent_id, account_type, is_detail, cost_center_id, active;

-- name: ListCostCenters :many
SELECT id, tenant_id, code, name, parent_id, active
FROM erp_cost_centers
WHERE tenant_id = $1 AND (sqlc.arg(active_only)::BOOLEAN = false OR active = true)
ORDER BY code;

-- name: CreateCostCenter :one
INSERT INTO erp_cost_centers (tenant_id, code, name, parent_id)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, code, name, parent_id, active;

-- name: ListFiscalYears :many
SELECT id, tenant_id, year, start_date, end_date, status
FROM erp_fiscal_years WHERE tenant_id = $1 ORDER BY year DESC;

-- name: CreateFiscalYear :one
INSERT INTO erp_fiscal_years (tenant_id, year, start_date, end_date)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, year, start_date, end_date, status;

-- name: ListJournalEntries :many
SELECT id, tenant_id, number, date, fiscal_year_id, concept, entry_type,
       reference_type, reference_id, user_id, status, created_at
FROM erp_journal_entries
WHERE tenant_id = $1
  AND (sqlc.arg(date_from)::DATE IS NULL OR date >= sqlc.arg(date_from)::DATE)
  AND (sqlc.arg(date_to)::DATE IS NULL OR date <= sqlc.arg(date_to)::DATE)
  AND (sqlc.arg(status_filter)::TEXT = '' OR status = sqlc.arg(status_filter)::TEXT)
ORDER BY date DESC, number DESC
LIMIT $2 OFFSET $3;

-- name: GetJournalEntry :one
SELECT id, tenant_id, number, date, fiscal_year_id, concept, entry_type,
       reference_type, reference_id, user_id, status, created_at
FROM erp_journal_entries WHERE id = $1 AND tenant_id = $2;

-- name: CreateJournalEntry :one
INSERT INTO erp_journal_entries (tenant_id, number, date, fiscal_year_id, concept, entry_type,
    reference_type, reference_id, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, tenant_id, number, date, fiscal_year_id, concept, entry_type,
    reference_type, reference_id, user_id, status, created_at;

-- name: PostJournalEntry :execrows
UPDATE erp_journal_entries SET status = 'posted'
WHERE id = $1 AND tenant_id = $2 AND status = 'draft';

-- name: ListJournalLines :many
SELECT jl.id, jl.tenant_id, jl.entry_id, jl.account_id, jl.cost_center_id,
       jl.entry_date, jl.debit, jl.credit, jl.description, jl.sort_order,
       a.code AS account_code, a.name AS account_name
FROM erp_journal_lines jl
JOIN erp_accounts a ON a.id = jl.account_id
WHERE jl.entry_id = $1 AND jl.tenant_id = $2
ORDER BY jl.sort_order;

-- name: CreateJournalLine :one
INSERT INTO erp_journal_lines (tenant_id, entry_id, account_id, cost_center_id, entry_date, debit, credit, description, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, tenant_id, entry_id, account_id, cost_center_id, entry_date, debit, credit, description, sort_order;

-- name: GetAccountBalance :many
SELECT jl.account_id, a.code AS account_code, a.name AS account_name,
       SUM(jl.debit)::NUMERIC(16,2) AS total_debit,
       SUM(jl.credit)::NUMERIC(16,2) AS total_credit,
       (SUM(jl.debit) - SUM(jl.credit))::NUMERIC(16,2) AS balance
FROM erp_journal_lines jl
JOIN erp_accounts a ON a.id = jl.account_id
JOIN erp_journal_entries je ON je.id = jl.entry_id
WHERE jl.tenant_id = $1 AND je.status = 'posted'
  AND (sqlc.arg(date_from)::DATE IS NULL OR jl.entry_date >= sqlc.arg(date_from)::DATE)
  AND (sqlc.arg(date_to)::DATE IS NULL OR jl.entry_date <= sqlc.arg(date_to)::DATE)
GROUP BY jl.account_id, a.code, a.name
ORDER BY a.code;

-- name: GetLedger :many
SELECT jl.id, jl.entry_date, jl.debit, jl.credit, jl.description,
       je.number AS entry_number, je.concept
FROM erp_journal_lines jl
JOIN erp_journal_entries je ON je.id = jl.entry_id
WHERE jl.tenant_id = $1 AND jl.account_id = $2 AND je.status = 'posted'
  AND (sqlc.arg(date_from)::DATE IS NULL OR jl.entry_date >= sqlc.arg(date_from)::DATE)
  AND (sqlc.arg(date_to)::DATE IS NULL OR jl.entry_date <= sqlc.arg(date_to)::DATE)
ORDER BY jl.entry_date, je.number
LIMIT $3 OFFSET $4;
