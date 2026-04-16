-- name: ListInvoices :many
SELECT i.id, i.tenant_id, i.number, i.date, i.due_date, i.invoice_type, i.direction,
       i.entity_id, i.currency_id, i.subtotal, i.tax_amount, i.total, i.order_id,
       i.afip_cae, i.status, i.user_id, i.created_at,
       e.name AS entity_name
FROM erp_invoices i
JOIN erp_entities e ON e.id = i.entity_id
WHERE i.tenant_id = $1
  AND (sqlc.arg(type_filter)::TEXT = '' OR i.invoice_type = sqlc.arg(type_filter)::TEXT)
  AND (sqlc.arg(direction_filter)::TEXT = '' OR i.direction = sqlc.arg(direction_filter)::TEXT)
  AND (sqlc.arg(status_filter)::TEXT = '' OR i.status = sqlc.arg(status_filter)::TEXT)
  AND (sqlc.arg(date_from)::DATE IS NULL OR i.date >= sqlc.arg(date_from)::DATE)
  AND (sqlc.arg(date_to)::DATE IS NULL OR i.date <= sqlc.arg(date_to)::DATE)
ORDER BY i.date DESC, i.number DESC
LIMIT $2 OFFSET $3;

-- name: GetInvoice :one
SELECT id, tenant_id, number, date, due_date, invoice_type, direction,
       entity_id, currency_id, subtotal, tax_amount, total, order_id,
       journal_entry_id, afip_cae, afip_cae_due, status, user_id, created_at
FROM erp_invoices WHERE id = $1 AND tenant_id = $2;

-- name: CreateInvoice :one
INSERT INTO erp_invoices (tenant_id, number, date, due_date, invoice_type, direction,
    entity_id, currency_id, subtotal, tax_amount, total, order_id, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING id, tenant_id, number, date, due_date, invoice_type, direction,
    entity_id, currency_id, subtotal, tax_amount, total, order_id,
    journal_entry_id, afip_cae, afip_cae_due, status, user_id, created_at;

-- name: PostInvoice :execrows
UPDATE erp_invoices SET status = 'posted'
WHERE id = $1 AND tenant_id = $2 AND status = 'draft';

-- name: ListInvoiceLines :many
SELECT il.id, il.tenant_id, il.invoice_id, il.article_id, il.description,
       il.quantity, il.unit_price, il.tax_rate, il.tax_amount, il.line_total, il.sort_order
FROM erp_invoice_lines il
WHERE il.invoice_id = $1 AND il.tenant_id = $2
ORDER BY il.sort_order;

-- name: CreateInvoiceLine :one
INSERT INTO erp_invoice_lines (tenant_id, invoice_id, article_id, description,
    quantity, unit_price, tax_rate, tax_amount, line_total, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, tenant_id, invoice_id, article_id, description,
    quantity, unit_price, tax_rate, tax_amount, line_total, sort_order;

-- name: GetTaxBook :many
SELECT te.id, te.tenant_id, te.invoice_id, te.period, te.direction,
       te.net_amount, te.tax_rate, te.tax_amount, te.created_at,
       i.number AS invoice_number, e.name AS entity_name
FROM erp_tax_entries te
JOIN erp_invoices i ON i.id = te.invoice_id
JOIN erp_entities e ON e.id = i.entity_id
WHERE te.tenant_id = $1 AND te.period = $2
  AND (sqlc.arg(direction_filter)::TEXT = '' OR te.direction = sqlc.arg(direction_filter)::TEXT)
ORDER BY i.date, i.number;

-- name: CreateTaxEntry :one
INSERT INTO erp_tax_entries (tenant_id, invoice_id, period, direction, net_amount, tax_rate, tax_amount)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, invoice_id, period, direction, net_amount, tax_rate, tax_amount, created_at;

-- name: ListWithholdings :many
SELECT w.id, w.tenant_id, w.invoice_id, w.movement_id, w.entity_id, w.type,
       w.rate, w.base_amount, w.amount, w.certificate_num, w.date, w.created_at,
       e.name AS entity_name
FROM erp_withholdings w
JOIN erp_entities e ON e.id = w.entity_id
WHERE w.tenant_id = $1
  AND (sqlc.arg(type_filter)::TEXT = '' OR w.type = sqlc.arg(type_filter)::TEXT)
ORDER BY w.date DESC
LIMIT $2 OFFSET $3;

-- ============================================================
-- Cascade void queries (Plan 18 Fase 2)
-- ============================================================

-- name: ListTaxEntriesByInvoice :many
SELECT id, tenant_id, invoice_id, period, direction, net_amount, tax_rate, tax_amount
FROM erp_tax_entries WHERE tenant_id = $1 AND invoice_id = $2;

-- name: VoidInvoice :execrows
UPDATE erp_invoices SET status = 'cancelled', voided_by = $3, void_reason = $4
WHERE id = $1 AND tenant_id = $2 AND status IN ('posted', 'paid');

-- name: CreateWithholding :one
INSERT INTO erp_withholdings (tenant_id, invoice_id, movement_id, entity_id, type, rate, base_amount, amount, certificate_num, date)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, tenant_id, invoice_id, movement_id, entity_id, type, rate, base_amount, amount, certificate_num, date, created_at;
