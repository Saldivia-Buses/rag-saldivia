-- name: ListAccountMovements :many
SELECT am.id, am.tenant_id, am.entity_id, am.date, am.movement_type, am.direction,
       am.amount, am.balance, am.invoice_id, am.treasury_id, am.notes, am.user_id, am.created_at,
       e.name AS entity_name
FROM erp_account_movements am
JOIN erp_entities e ON e.id = am.entity_id
WHERE am.tenant_id = $1
  AND (sqlc.arg(entity_filter)::UUID IS NULL OR am.entity_id = sqlc.arg(entity_filter)::UUID)
  AND (sqlc.arg(direction_filter)::TEXT = '' OR am.direction = sqlc.arg(direction_filter)::TEXT)
  AND (sqlc.arg(date_from)::DATE IS NULL OR am.date >= sqlc.arg(date_from)::DATE)
  AND (sqlc.arg(date_to)::DATE IS NULL OR am.date <= sqlc.arg(date_to)::DATE)
ORDER BY am.date DESC, am.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateAccountMovement :one
INSERT INTO erp_account_movements (tenant_id, entity_id, date, movement_type, direction,
    amount, balance, invoice_id, treasury_id, journal_entry_id, notes, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING id, tenant_id, entity_id, date, movement_type, direction, amount, balance,
    invoice_id, treasury_id, journal_entry_id, notes, user_id, created_at;

-- name: GetEntityBalances :many
SELECT am.entity_id, e.name AS entity_name, e.type AS entity_type, am.direction,
       SUM(am.balance)::NUMERIC(16,2) AS open_balance
FROM erp_account_movements am
JOIN erp_entities e ON e.id = am.entity_id
WHERE am.tenant_id = $1 AND am.balance > 0
  AND (sqlc.arg(direction_filter)::TEXT = '' OR am.direction = sqlc.arg(direction_filter)::TEXT)
GROUP BY am.entity_id, e.name, e.type, am.direction
ORDER BY open_balance DESC;

-- name: GetOverdueInvoices :many
SELECT am.id, am.entity_id, am.date, am.amount, am.balance, am.direction, am.created_at,
       e.name AS entity_name, i.number AS invoice_number, i.due_date
FROM erp_account_movements am
JOIN erp_entities e ON e.id = am.entity_id
LEFT JOIN erp_invoices i ON i.id = am.invoice_id
WHERE am.tenant_id = $1 AND am.balance > 0 AND am.movement_type = 'invoice'
  AND i.due_date IS NOT NULL AND i.due_date < CURRENT_DATE
ORDER BY i.due_date;

-- name: CreatePaymentAllocation :one
INSERT INTO erp_payment_allocations (tenant_id, payment_id, invoice_id, amount)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, payment_id, invoice_id, amount, created_at;

-- name: UpdateMovementBalance :exec
UPDATE erp_account_movements SET balance = balance - $3
WHERE id = $1 AND tenant_id = $2 AND balance >= $3;
