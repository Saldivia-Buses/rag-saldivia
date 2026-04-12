-- name: ListArticles :many
SELECT id, tenant_id, code, name, family_id, category_id, unit_id, article_type,
       min_stock, max_stock, reorder_point, last_cost, avg_cost, metadata, active, created_at, updated_at
FROM erp_articles
WHERE tenant_id = $1
  AND (sqlc.arg(active_only)::BOOLEAN = false OR active = true)
  AND (sqlc.arg(search)::TEXT = '' OR name ILIKE '%' || sqlc.arg(search)::TEXT || '%'
       OR code ILIKE '%' || sqlc.arg(search)::TEXT || '%')
  AND (sqlc.arg(article_type_filter)::TEXT = '' OR article_type = sqlc.arg(article_type_filter)::TEXT)
ORDER BY code
LIMIT $2 OFFSET $3;

-- name: GetArticle :one
SELECT id, tenant_id, code, name, family_id, category_id, unit_id, article_type,
       min_stock, max_stock, reorder_point, last_cost, avg_cost, metadata, active, created_at, updated_at
FROM erp_articles WHERE id = $1 AND tenant_id = $2;

-- name: CreateArticle :one
INSERT INTO erp_articles (tenant_id, code, name, family_id, category_id, unit_id, article_type,
    min_stock, max_stock, reorder_point, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, tenant_id, code, name, family_id, category_id, unit_id, article_type,
    min_stock, max_stock, reorder_point, last_cost, avg_cost, metadata, active, created_at, updated_at;

-- name: UpdateArticle :one
UPDATE erp_articles
SET code = $3, name = $4, family_id = $5, category_id = $6, unit_id = $7, article_type = $8,
    min_stock = $9, max_stock = $10, reorder_point = $11, metadata = $12, active = $13, updated_at = now()
WHERE id = $1 AND tenant_id = $2
RETURNING id, tenant_id, code, name, family_id, category_id, unit_id, article_type,
    min_stock, max_stock, reorder_point, last_cost, avg_cost, metadata, active, created_at, updated_at;

-- name: ListWarehouses :many
SELECT id, tenant_id, code, name, location, active
FROM erp_warehouses WHERE tenant_id = $1 AND (sqlc.arg(active_only)::BOOLEAN = false OR active = true)
ORDER BY code;

-- name: CreateWarehouse :one
INSERT INTO erp_warehouses (tenant_id, code, name, location)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, code, name, location, active;

-- name: GetStockLevels :many
SELECT sl.id, sl.tenant_id, sl.article_id, sl.warehouse_id, sl.quantity, sl.reserved, sl.updated_at,
       a.code AS article_code, a.name AS article_name, w.code AS warehouse_code, w.name AS warehouse_name
FROM erp_stock_levels sl
JOIN erp_articles a ON a.id = sl.article_id
JOIN erp_warehouses w ON w.id = sl.warehouse_id
WHERE sl.tenant_id = $1
  AND (sqlc.arg(article_filter)::UUID IS NULL OR sl.article_id = sqlc.arg(article_filter)::UUID)
  AND (sqlc.arg(warehouse_filter)::UUID IS NULL OR sl.warehouse_id = sqlc.arg(warehouse_filter)::UUID)
ORDER BY a.code, w.code;

-- name: ListStockMovements :many
SELECT sm.id, sm.tenant_id, sm.article_id, sm.warehouse_id, sm.movement_type, sm.quantity,
       sm.unit_cost, sm.reference_type, sm.reference_id, sm.concept_id, sm.user_id, sm.notes, sm.created_at,
       a.code AS article_code, a.name AS article_name
FROM erp_stock_movements sm
JOIN erp_articles a ON a.id = sm.article_id
WHERE sm.tenant_id = $1
  AND (sqlc.arg(article_filter)::UUID IS NULL OR sm.article_id = sqlc.arg(article_filter)::UUID)
ORDER BY sm.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateStockMovement :one
INSERT INTO erp_stock_movements (tenant_id, article_id, warehouse_id, movement_type, quantity, unit_cost,
    reference_type, reference_id, concept_id, user_id, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, tenant_id, article_id, warehouse_id, movement_type, quantity, unit_cost,
    reference_type, reference_id, concept_id, user_id, notes, created_at;

-- name: UpsertStockLevel :exec
INSERT INTO erp_stock_levels (tenant_id, article_id, warehouse_id, quantity, updated_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (tenant_id, article_id, warehouse_id)
DO UPDATE SET quantity = erp_stock_levels.quantity + $4, updated_at = now();

-- name: ListStockMovementsByRef :many
SELECT id, tenant_id, article_id, warehouse_id, movement_type, quantity, unit_cost,
       reference_type, reference_id
FROM erp_stock_movements
WHERE tenant_id = $1 AND reference_type = $2 AND reference_id = $3;

-- name: ListBOM :many
SELECT b.id, b.tenant_id, b.parent_id, b.child_id, b.quantity, b.unit_id, b.sort_order, b.notes,
       a.code AS child_code, a.name AS child_name
FROM erp_bom b
JOIN erp_articles a ON a.id = b.child_id
WHERE b.tenant_id = $1 AND b.parent_id = $2
ORDER BY b.sort_order, a.name;

-- name: CreateBOMEntry :one
INSERT INTO erp_bom (tenant_id, parent_id, child_id, quantity, unit_id, sort_order, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, parent_id, child_id, quantity, unit_id, sort_order, notes;

-- name: DeleteBOMEntry :exec
DELETE FROM erp_bom WHERE id = $1 AND tenant_id = $2;
