-- name: ListProductionCenters :many
SELECT id, tenant_id, code, name, active
FROM erp_production_centers WHERE tenant_id = $1 ORDER BY code;

-- name: CreateProductionCenter :one
INSERT INTO erp_production_centers (tenant_id, code, name)
VALUES ($1, $2, $3)
RETURNING id, tenant_id, code, name, active;

-- name: ListProductionOrders :many
SELECT po.id, po.tenant_id, po.number, po.date, po.product_id, po.center_id,
       po.quantity, po.status, po.priority, po.order_id, po.start_date, po.end_date,
       po.user_id, po.notes, po.created_at,
       a.code AS product_code, a.name AS product_name
FROM erp_production_orders po
JOIN erp_articles a ON a.id = po.product_id
WHERE po.tenant_id = $1
  AND (sqlc.arg(status_filter)::TEXT = '' OR po.status = sqlc.arg(status_filter)::TEXT)
ORDER BY po.date DESC
LIMIT $2 OFFSET $3;

-- name: GetProductionOrder :one
SELECT id, tenant_id, number, date, product_id, center_id, quantity, status,
       priority, order_id, start_date, end_date, user_id, notes, created_at
FROM erp_production_orders WHERE id = $1 AND tenant_id = $2;

-- name: CreateProductionOrder :one
INSERT INTO erp_production_orders (tenant_id, number, date, product_id, center_id, quantity,
    priority, order_id, start_date, end_date, user_id, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING id, tenant_id, number, date, product_id, center_id, quantity, status,
    priority, order_id, start_date, end_date, user_id, notes, created_at;

-- name: UpdateProductionOrderStatus :execrows
UPDATE erp_production_orders SET status = $3, start_date = COALESCE(start_date, CURRENT_DATE)
WHERE id = $1 AND tenant_id = $2;

-- name: ListProductionMaterials :many
SELECT pm.id, pm.tenant_id, pm.order_id, pm.article_id, pm.required_qty,
       pm.consumed_qty, pm.warehouse_id,
       a.code AS article_code, a.name AS article_name
FROM erp_production_materials pm
JOIN erp_articles a ON a.id = pm.article_id
WHERE pm.order_id = $1 AND pm.tenant_id = $2
ORDER BY a.code;

-- name: CreateProductionMaterial :one
INSERT INTO erp_production_materials (tenant_id, order_id, article_id, required_qty, warehouse_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, order_id, article_id, required_qty, consumed_qty, warehouse_id;

-- name: ListProductionSteps :many
SELECT id, tenant_id, order_id, step_name, sort_order, status,
       assigned_to, started_at, completed_at, notes
FROM erp_production_steps WHERE order_id = $1 AND tenant_id = $2
ORDER BY sort_order;

-- name: CreateProductionStep :one
INSERT INTO erp_production_steps (tenant_id, order_id, step_name, sort_order, assigned_to)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, order_id, step_name, sort_order, status, assigned_to, started_at, completed_at, notes;

-- name: UpdateProductionStep :execrows
UPDATE erp_production_steps SET status = $3, started_at = CASE WHEN $3 = 'in_progress' AND started_at IS NULL THEN now() ELSE started_at END,
    completed_at = CASE WHEN $3 = 'completed' THEN now() ELSE completed_at END, notes = COALESCE(NULLIF($4, ''), notes)
WHERE id = $1 AND tenant_id = $2;

-- name: CreateProductionInspection :one
INSERT INTO erp_production_inspections (tenant_id, order_id, step_id, inspector_id, result, observations)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tenant_id, order_id, step_id, inspector_id, result, observations, created_at;

-- name: ListUnits :many
SELECT id, tenant_id, chassis_number, internal_number, model, customer_id,
       order_id, production_order_id, patent, status, engine_brand, body_style,
       seat_count, year, delivered_at, created_at
FROM erp_units WHERE tenant_id = $1
  AND (sqlc.arg(status_filter)::TEXT = '' OR status = sqlc.arg(status_filter)::TEXT)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUnit :one
SELECT id, tenant_id, chassis_number, internal_number, model, customer_id,
       order_id, production_order_id, patent, status, engine_brand, body_style,
       seat_count, year, metadata, delivered_at, created_at
FROM erp_units WHERE id = $1 AND tenant_id = $2;

-- name: CreateUnit :one
INSERT INTO erp_units (tenant_id, chassis_number, internal_number, model, customer_id,
    order_id, production_order_id, patent, engine_brand, body_style, seat_count, year, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING id, tenant_id, chassis_number, internal_number, model, customer_id,
    order_id, production_order_id, patent, status, engine_brand, body_style,
    seat_count, year, metadata, delivered_at, created_at;

-- name: UpdateUnitStatus :execrows
UPDATE erp_units SET status = $3, delivered_at = CASE WHEN $3 = 'delivered' THEN CURRENT_DATE ELSE delivered_at END
WHERE id = $1 AND tenant_id = $2;
