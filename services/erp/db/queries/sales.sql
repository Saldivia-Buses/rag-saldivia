-- name: ListQuotations :many
SELECT q.id, q.tenant_id, q.number, q.date, q.customer_id, q.status,
       q.currency_id, q.total, q.valid_until, q.notes, q.user_id, q.created_at,
       e.name AS customer_name
FROM erp_quotations q
JOIN erp_entities e ON e.id = q.customer_id
WHERE q.tenant_id = $1
  AND (sqlc.arg(status_filter)::TEXT = '' OR q.status = sqlc.arg(status_filter)::TEXT)
ORDER BY q.date DESC
LIMIT $2 OFFSET $3;

-- name: GetQuotation :one
SELECT id, tenant_id, number, date, customer_id, status, currency_id, total,
       valid_until, notes, user_id, created_at
FROM erp_quotations WHERE id = $1 AND tenant_id = $2;

-- name: CreateQuotation :one
INSERT INTO erp_quotations (tenant_id, number, date, customer_id, currency_id, total, valid_until, notes, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, tenant_id, number, date, customer_id, status, currency_id, total, valid_until, notes, user_id, created_at;

-- name: UpdateQuotationStatus :execrows
UPDATE erp_quotations SET status = $3 WHERE id = $1 AND tenant_id = $2;

-- name: ListQuotationLines :many
SELECT ql.id, ql.tenant_id, ql.quotation_id, ql.article_id, ql.description,
       ql.quantity, ql.unit_price, ql.sort_order, ql.metadata
FROM erp_quotation_lines ql
WHERE ql.quotation_id = $1 AND ql.tenant_id = $2
ORDER BY ql.sort_order;

-- name: CreateQuotationLine :one
INSERT INTO erp_quotation_lines (tenant_id, quotation_id, article_id, description, quantity, unit_price, sort_order, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, tenant_id, quotation_id, article_id, description, quantity, unit_price, sort_order, metadata;

-- name: ListOrders :many
SELECT o.id, o.tenant_id, o.number, o.date, o.order_type, o.customer_id,
       o.quotation_id, o.status, o.total, o.user_id, o.notes, o.created_at,
       e.name AS customer_name
FROM erp_orders o
LEFT JOIN erp_entities e ON e.id = o.customer_id
WHERE o.tenant_id = $1
  AND (sqlc.arg(status_filter)::TEXT = '' OR o.status = sqlc.arg(status_filter)::TEXT)
  AND (sqlc.arg(type_filter)::TEXT = '' OR o.order_type = sqlc.arg(type_filter)::TEXT)
ORDER BY o.date DESC
LIMIT $2 OFFSET $3;

-- name: GetOrder :one
SELECT id, tenant_id, number, date, order_type, customer_id, quotation_id,
       status, total, user_id, notes, created_at
FROM erp_orders WHERE id = $1 AND tenant_id = $2;

-- name: CreateOrder :one
INSERT INTO erp_orders (tenant_id, number, date, order_type, customer_id, quotation_id, total, user_id, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, tenant_id, number, date, order_type, customer_id, quotation_id, status, total, user_id, notes, created_at;

-- name: UpdateOrderStatus :execrows
UPDATE erp_orders SET status = $3 WHERE id = $1 AND tenant_id = $2;

-- name: ListPriceLists :many
SELECT id, tenant_id, name, currency_id, valid_from, valid_until, active
FROM erp_price_lists WHERE tenant_id = $1 ORDER BY name;

-- name: CreatePriceList :one
INSERT INTO erp_price_lists (tenant_id, name, currency_id, valid_from, valid_until)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, name, currency_id, valid_from, valid_until, active;

-- name: ListPriceListItems :many
SELECT pli.id, pli.tenant_id, pli.price_list_id, pli.article_id, pli.description, pli.price,
       a.code AS article_code, a.name AS article_name
FROM erp_price_list_items pli
LEFT JOIN erp_articles a ON a.id = pli.article_id
WHERE pli.price_list_id = $1 AND pli.tenant_id = $2
ORDER BY a.code, pli.description;

-- name: CreatePriceListItem :one
INSERT INTO erp_price_list_items (tenant_id, price_list_id, article_id, description, price)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, price_list_id, article_id, description, price;
