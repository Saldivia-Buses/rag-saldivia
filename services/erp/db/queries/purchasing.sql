-- name: ListPurchaseOrders :many
SELECT po.id, po.tenant_id, po.number, po.date, po.supplier_id, po.status,
       po.currency_id, po.total, po.notes, po.user_id, po.created_at,
       e.name AS supplier_name
FROM erp_purchase_orders po
JOIN erp_entities e ON e.id = po.supplier_id
WHERE po.tenant_id = $1
  AND (sqlc.arg(status_filter)::TEXT = '' OR po.status = sqlc.arg(status_filter)::TEXT)
ORDER BY po.date DESC, po.number DESC
LIMIT $2 OFFSET $3;

-- name: GetPurchaseOrder :one
SELECT id, tenant_id, number, date, supplier_id, status, currency_id, total, notes, user_id, created_at
FROM erp_purchase_orders WHERE id = $1 AND tenant_id = $2;

-- name: CreatePurchaseOrder :one
INSERT INTO erp_purchase_orders (tenant_id, number, date, supplier_id, currency_id, total, notes, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, tenant_id, number, date, supplier_id, status, currency_id, total, notes, user_id, created_at;

-- name: ApprovePurchaseOrder :execrows
UPDATE erp_purchase_orders SET status = 'approved'
WHERE id = $1 AND tenant_id = $2 AND status = 'draft';

-- name: ListPurchaseOrderLines :many
SELECT pol.id, pol.tenant_id, pol.order_id, pol.article_id, pol.quantity,
       pol.unit_price, pol.received_qty, pol.sort_order,
       a.code AS article_code, a.name AS article_name
FROM erp_purchase_order_lines pol
JOIN erp_articles a ON a.id = pol.article_id
WHERE pol.order_id = $1 AND pol.tenant_id = $2
ORDER BY pol.sort_order;

-- name: CreatePurchaseOrderLine :one
INSERT INTO erp_purchase_order_lines (tenant_id, order_id, article_id, quantity, unit_price, sort_order)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tenant_id, order_id, article_id, quantity, unit_price, received_qty, sort_order;

-- name: CreatePurchaseReceipt :one
INSERT INTO erp_purchase_receipts (tenant_id, order_id, date, number, user_id, notes)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tenant_id, order_id, date, number, user_id, notes, created_at;

-- name: CreatePurchaseReceiptLine :one
INSERT INTO erp_purchase_receipt_lines (tenant_id, receipt_id, order_line_id, article_id, quantity)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, receipt_id, order_line_id, article_id, quantity;

-- name: UpdateReceivedQty :exec
UPDATE erp_purchase_order_lines SET received_qty = received_qty + $3
WHERE id = $1 AND tenant_id = $2;

-- name: GetPurchaseReceipt :one
SELECT pr.id, pr.tenant_id, pr.order_id, pr.date, pr.number, pr.user_id, pr.notes, pr.created_at
FROM erp_purchase_receipts pr
WHERE pr.id = $1 AND pr.tenant_id = $2;

-- ============================================================
-- QC Inspection queries (Plan 18 Fase 3)
-- ============================================================

-- name: CreateInspection :one
INSERT INTO erp_qc_inspections (tenant_id, receipt_id, receipt_line_id, article_id,
    quantity, accepted_qty, rejected_qty, status, inspector_id, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, tenant_id, receipt_id, receipt_line_id, article_id, quantity,
    accepted_qty, rejected_qty, status, inspector_id, notes, completed_at, created_at;

-- name: CompleteInspection :execrows
UPDATE erp_qc_inspections SET status = 'completed', completed_at = now()
WHERE id = $1 AND tenant_id = $2 AND status = 'pending';

-- name: ListInspections :many
SELECT qi.id, qi.tenant_id, qi.receipt_id, qi.article_id, qi.quantity,
       qi.accepted_qty, qi.rejected_qty, qi.status, qi.inspector_id, qi.created_at,
       a.code AS article_code, a.name AS article_name, pr.number AS receipt_number
FROM erp_qc_inspections qi
JOIN erp_articles a ON a.id = qi.article_id
JOIN erp_purchase_receipts pr ON pr.id = qi.receipt_id
WHERE qi.tenant_id = $1
  AND (sqlc.arg(status_filter)::TEXT = '' OR qi.status = sqlc.arg(status_filter)::TEXT)
ORDER BY qi.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetInspection :one
SELECT qi.id, qi.tenant_id, qi.receipt_id, qi.receipt_line_id, qi.article_id,
       qi.quantity, qi.accepted_qty, qi.rejected_qty, qi.status, qi.inspector_id,
       qi.notes, qi.completed_at, qi.created_at
FROM erp_qc_inspections qi WHERE qi.id = $1 AND qi.tenant_id = $2;

-- name: CreateDemerit :one
INSERT INTO erp_supplier_demerits (tenant_id, supplier_id, inspection_id, points, reason)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, supplier_id, inspection_id, points, reason, created_at;

-- name: ListSupplierDemerits :many
SELECT sd.id, sd.tenant_id, sd.supplier_id, sd.inspection_id, sd.points, sd.reason, sd.created_at
FROM erp_supplier_demerits sd
WHERE sd.tenant_id = $1 AND sd.supplier_id = $2
ORDER BY sd.created_at DESC;

-- name: GetSupplierDemeritTotal :one
SELECT COALESCE(SUM(points), 0)::INT AS total_points
FROM erp_supplier_demerits WHERE tenant_id = $1 AND supplier_id = $2;

-- name: ListPurchaseReceiptLines :many
SELECT prl.id, prl.tenant_id, prl.receipt_id, prl.order_line_id, prl.article_id, prl.quantity,
       a.code AS article_code, a.name AS article_name
FROM erp_purchase_receipt_lines prl
JOIN erp_articles a ON a.id = prl.article_id
WHERE prl.tenant_id = $1 AND prl.receipt_id = $2
ORDER BY prl.id;

-- name: ListPurchaseReceipts :many
SELECT pr.id, pr.tenant_id, pr.order_id, pr.date, pr.number, pr.user_id, pr.notes, pr.created_at,
       po.number AS order_number
FROM erp_purchase_receipts pr
JOIN erp_purchase_orders po ON po.id = pr.order_id
WHERE pr.tenant_id = $1
ORDER BY pr.date DESC
LIMIT $2 OFFSET $3;
