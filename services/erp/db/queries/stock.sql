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

-- name: ListBOMHistory :many
-- Historical BOM entries migrated from Histrix STK_BOM_HIST (14.7M rows on
-- saldivia_bench). Filter by parent article so the UI can render the
-- evolution of a single assembly's bill of materials without pulling the
-- entire history table.
SELECT h.id, h.tenant_id, h.parent_id, h.child_id, h.quantity, h.unit_id,
       h.version, h.effective_date, h.replaced_date, h.legacy_id, h.created_at,
       c.code AS child_code, c.name AS child_name
FROM erp_bom_history h
JOIN erp_articles c ON c.id = h.child_id
WHERE h.tenant_id = $1 AND h.parent_id = $2
ORDER BY h.effective_date DESC, h.version DESC
LIMIT $3 OFFSET $4;

-- ─── Per-supplier article costs (Phase 1 §Data migration — Pareto #5) ──────

-- name: ListArticleCosts :many
-- Per-supplier cost ledger for a single article. The migrated STKINSPR
-- rows carry one entry per (article, supplier) pair with the last update
-- date; this query lets the UI "costos" screens replace the Histrix
-- stock/costos/ views 1:1.
SELECT id, tenant_id, legacy_id, article_code, article_id, subsystem_code,
       cost, percentage_1, percentage_2, percentage_3,
       supplier_article_code, supplier_code, supplier_entity_id,
       invoice_date, last_update_date,
       movement_no, movement_post, movement_date, recalc_flag, created_at
FROM erp_article_costs
WHERE tenant_id = $1 AND article_id = $2
ORDER BY last_update_date DESC NULLS LAST, supplier_code
LIMIT $3 OFFSET $4;

-- name: ListArticleCostHistory :many
-- Monthly cost history snapshots per article (STK_COSTO_HIST migrated).
-- Most-recent period first. Backs the evolucion_costos / evolutivo_costo
-- screens 1:1 post-cutover.
SELECT id, tenant_id, legacy_id, article_code, article_id,
       year, month, cost, period_code, created_at
FROM erp_article_cost_history
WHERE tenant_id = $1 AND article_id = $2
ORDER BY year DESC, month DESC
LIMIT $3 OFFSET $4;

-- ─── Article replacement cost history (STK_COSTO_REPOSICION_HIST migrated — 2.0.11) ───

-- name: ListArticleReplacementCostHistory :many
-- Rolling log of supplier replacement-cost changes. Filter by supplier
-- entity or modified-at range. Feeds the stock cost-evolution views.
SELECT id, tenant_id, legacy_id, replacement_cost_legacy_id,
       supplier_entity_id, supplier_legacy_id,
       currency_id, currency_legacy_id,
       exchange_rate, supplier_cost,
       origin, incoterm,
       import_expenses, local_freight,
       modified_at, discount_1, discount_2, created_at
FROM erp_article_replacement_cost_history
WHERE tenant_id = $1
  AND (sqlc.arg(supplier_filter)::UUID IS NULL OR supplier_entity_id = sqlc.arg(supplier_filter)::UUID)
  AND (sqlc.arg(date_from)::TIMESTAMPTZ IS NULL OR modified_at >= sqlc.arg(date_from)::TIMESTAMPTZ)
  AND (sqlc.arg(date_to)::TIMESTAMPTZ IS NULL OR modified_at <= sqlc.arg(date_to)::TIMESTAMPTZ)
ORDER BY modified_at DESC NULLS LAST, legacy_id DESC
LIMIT $2 OFFSET $3;

-- ─── Stock cost movements (STK_COSTOS migrated — 2.0.12) ───

-- name: ListStockCostMovements :many
-- Priced stock-movement ledger. Filters by article, entity, deposit,
-- or date range. Feeds presup/* + estadisticas/evolutivo_costo +
-- stock_local/stkinmov_ingresos views.
SELECT id, tenant_id, legacy_id,
       article_code, article_id,
       entity_legacy_id, entity_id, account_legacy_code,
       deposit_legacy_id, sector_legacy_id, family_legacy_id,
       rubro_legacy_id, list_legacy_id, concept_legacy_id,
       unit_legacy_id, subsystem_code,
       movement_date, registered_date, invoice_date,
       station, movement_no, movement_order,
       reference, barcode, description,
       title_code, operator_class, operator_code,
       register_min, branch_code, unit_type,
       quantity, cost_price, sale_price, total_price, average_price,
       bonus_pct, purchase_amount, pending_amount,
       peso_amount, usage_amount, sale_ref,
       chassis_no, order_cps_no,
       invoice_id, invoice_legacy_id, invoice_line_legacy_id,
       cps_movement_legacy_id, cps_detail_legacy_id,
       order_detail_legacy_id, cash_movement_legacy_id,
       order_legacy_id, user_legacy_id, created_at
FROM erp_stock_cost_movements
WHERE tenant_id = $1
  AND (sqlc.arg(article_filter)::UUID IS NULL OR article_id = sqlc.arg(article_filter)::UUID)
  AND (sqlc.arg(entity_filter)::UUID IS NULL OR entity_id = sqlc.arg(entity_filter)::UUID)
  AND (sqlc.arg(deposit_filter)::INTEGER = 0 OR deposit_legacy_id = sqlc.arg(deposit_filter)::INTEGER)
  AND (sqlc.arg(date_from)::DATE IS NULL OR movement_date >= sqlc.arg(date_from)::DATE)
  AND (sqlc.arg(date_to)::DATE IS NULL OR movement_date <= sqlc.arg(date_to)::DATE)
ORDER BY movement_date DESC NULLS LAST, legacy_id DESC
LIMIT $2 OFFSET $3;
