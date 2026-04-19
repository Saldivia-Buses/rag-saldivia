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

-- name: ListAccountMovementsByInvoice :many
SELECT id, tenant_id, entity_id, date, movement_type, direction, amount, balance,
       invoice_id, treasury_id, journal_entry_id, notes, user_id
FROM erp_account_movements
WHERE tenant_id = $1 AND invoice_id = $2;

-- name: CreatePaymentAllocation :one
INSERT INTO erp_payment_allocations (tenant_id, payment_id, invoice_id, amount)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, payment_id, invoice_id, amount, created_at;

-- name: UpdateMovementBalance :execrows
UPDATE erp_account_movements SET balance = balance - $3
WHERE id = $1 AND tenant_id = $2 AND balance >= $3;

-- ─── Invoice notes (REG_MOVIMIENTO_OBS migrated — 2.0.11) ───

-- name: ListInvoiceNotes :many
-- Free-text notes attached to REG_MOVIMIENTOS (now erp_invoices).
-- Filter by invoice_id or by tenant-wide date range. Mirrors
-- proveedores_loc/cliente observaciones.
SELECT id, tenant_id, legacy_id, observation_date, observation_time,
       observation, invoice_id, invoice_legacy_id,
       login, contact_legacy_id, source_table, system_code,
       movement_date, account_code, concept_code,
       movement_voucher_class, movement_no, created_at
FROM erp_invoice_notes
WHERE tenant_id = $1
  AND (sqlc.arg(invoice_filter)::UUID IS NULL OR invoice_id = sqlc.arg(invoice_filter)::UUID)
  AND (sqlc.arg(date_from)::DATE IS NULL OR observation_date >= sqlc.arg(date_from)::DATE)
  AND (sqlc.arg(date_to)::DATE IS NULL OR observation_date <= sqlc.arg(date_to)::DATE)
ORDER BY observation_date DESC NULLS LAST, legacy_id DESC
LIMIT $2 OFFSET $3;

-- ─── Payment complaints (RECLAMOPAGOS migrated — 2.0.11) ───

-- name: ListPaymentComplaints :many
-- Supplier-payment reclamation log. Filter by pending/done status or
-- entity. Mirrors the reclamos/reclamopagos.xml "abm-mini" view.
SELECT id, tenant_id, legacy_id, complaint_date,
       entity_legacy_code, entity_id,
       observation, status_flag, login, created_at
FROM erp_payment_complaints
WHERE tenant_id = $1
  AND (sqlc.arg(status_filter)::SMALLINT = -1 OR status_flag = sqlc.arg(status_filter)::SMALLINT)
  AND (sqlc.arg(entity_filter)::UUID IS NULL OR entity_id = sqlc.arg(entity_filter)::UUID)
ORDER BY complaint_date DESC NULLS LAST, legacy_id DESC
LIMIT $2 OFFSET $3;

-- name: CreatePaymentComplaint :one
-- Create a new supplier-payment complaint. entity_id resolves at the
-- caller (frontend picks the entity UUID).
INSERT INTO erp_payment_complaints (
    tenant_id, complaint_date, entity_id, entity_legacy_code,
    observation, status_flag, login
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, legacy_id, complaint_date,
          entity_legacy_code, entity_id, observation,
          status_flag, login, created_at;

-- name: UpdatePaymentComplaintStatus :execrows
-- Flip the marca flag (0=pendiente ↔ 1=cumplida) on a single complaint.
UPDATE erp_payment_complaints
   SET status_flag = $3
 WHERE id = $1 AND tenant_id = $2;

-- name: GetPaymentComplaint :one
SELECT id, tenant_id, legacy_id, complaint_date,
       entity_legacy_code, entity_id, observation,
       status_flag, login, created_at
  FROM erp_payment_complaints
 WHERE id = $1 AND tenant_id = $2;
