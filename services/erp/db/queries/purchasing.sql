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

-- name: ListPurchaseReceipts :many
SELECT pr.id, pr.tenant_id, pr.order_id, pr.date, pr.number, pr.user_id, pr.notes, pr.created_at,
       po.number AS order_number
FROM erp_purchase_receipts pr
JOIN erp_purchase_orders po ON po.id = pr.order_id
WHERE pr.tenant_id = $1
ORDER BY pr.date DESC
LIMIT $2 OFFSET $3;
