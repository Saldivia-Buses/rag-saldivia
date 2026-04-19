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

-- name: StartProductionOrder :execrows
UPDATE erp_production_orders SET status = 'in_progress', start_date = COALESCE(start_date, CURRENT_DATE)
WHERE id = $1 AND tenant_id = $2 AND status = 'planned';

-- name: CompleteProductionOrder :execrows
UPDATE erp_production_orders SET status = 'completed', end_date = COALESCE(end_date, CURRENT_DATE)
WHERE id = $1 AND tenant_id = $2 AND status = 'in_progress';

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

-- name: ListHomologations :many
SELECT id, tenant_id, plano, expte, dispos, fecha_aprob, fecha_vto,
       seats, seats_lower, weight_tare, weight_gross, vin,
       commercial_code, commercial_desc, active, created_at
FROM erp_homologations
WHERE tenant_id = $1
  AND (sqlc.arg(active_only)::BOOLEAN = false OR active = true)
ORDER BY commercial_code, id
LIMIT $2 OFFSET $3;

-- name: ListHomologationRevisions :many
SELECT id, tenant_id, homologation_id, date, notes, created_at
FROM erp_homologation_revisions
WHERE tenant_id = $1 AND homologation_id = $2
ORDER BY date DESC, id
LIMIT $3 OFFSET $4;

-- name: ListHomologationRevisionLines :many
SELECT id, tenant_id, revision_id, article_id, article_code, article_desc, article_unit,
       process_1, process_2, process_3, process_4,
       multiplier, quantity, replacement_cost, replacement_partial,
       replacement_cost_desc, replacement_partial_desc,
       account_code, account_name, partial_with_surcharge, region_percentage,
       partial_clog, partial_surcharge_log, logistics_cost, created_at
FROM erp_homologation_revision_lines
WHERE tenant_id = $1 AND revision_id = $2
ORDER BY process_1, process_2, process_3, process_4, article_code
LIMIT $3 OFFSET $4;

-- name: ListProductionInspectionHomologations :many
-- Production inspection templates × homologated vehicle models
-- (PROD_CONTROL_HOMOLOG migrated). Pareto #7. The UI at
-- controlcalidad/prod_control_homolog.xml iterates homologations for
-- a single inspection; this backs that view post-cutover.
SELECT id, tenant_id, legacy_id,
       inspection_id, inspection_legacy_id,
       homologation_id, homologation_legacy_id, created_at
FROM erp_production_inspection_homologations
WHERE tenant_id = $1 AND inspection_id = $2
ORDER BY homologation_legacy_id
LIMIT $3 OFFSET $4;

-- ─── Unit accessories (ACCESORIOS_COCHE migrated — 2.0.11) ───

-- name: ListUnitAccessories :many
-- Per-unit accessory lines. Filter by unit, order or date range.
-- Mirrors the vehicle-order accessory views in Histrix.
SELECT id, tenant_id, legacy_id,
       unit_id, unit_legacy_id,
       article_code, article_id, article_description,
       accessory_date,
       quotation_id, quotation_legacy_id,
       order_id, order_legacy_id,
       status, additional_price, quantity,
       approved_at, unit_price,
       product_section_id, product_section_legacy_id,
       observations, show_on_fv, show_on_ft,
       accessory_state_legacy_id, created_at
FROM erp_unit_accessories
WHERE tenant_id = $1
  AND (sqlc.arg(unit_filter)::UUID IS NULL OR unit_id = sqlc.arg(unit_filter)::UUID)
  AND (sqlc.arg(order_filter)::UUID IS NULL OR order_id = sqlc.arg(order_filter)::UUID)
  AND (sqlc.arg(date_from)::DATE IS NULL OR accessory_date >= sqlc.arg(date_from)::DATE)
  AND (sqlc.arg(date_to)::DATE IS NULL OR accessory_date <= sqlc.arg(date_to)::DATE)
ORDER BY accessory_date DESC NULLS LAST, legacy_id DESC
LIMIT $2 OFFSET $3;
