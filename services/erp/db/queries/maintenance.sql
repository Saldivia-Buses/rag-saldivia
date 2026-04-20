-- name: ListMaintenanceAssets :many
SELECT id, tenant_id, code, name, asset_type, unit_id, location, metadata, active, created_at
FROM erp_maintenance_assets WHERE tenant_id = $1
  AND (sqlc.arg(active_only)::BOOLEAN = false OR active = true)
  AND (sqlc.arg(type_filter)::TEXT = '' OR asset_type = sqlc.arg(type_filter)::TEXT)
ORDER BY code;

-- name: GetMaintenanceAsset :one
SELECT id, tenant_id, code, name, asset_type, unit_id, location, metadata, active, created_at
FROM erp_maintenance_assets
WHERE id = $1 AND tenant_id = $2;

-- name: CreateMaintenanceAsset :one
INSERT INTO erp_maintenance_assets (tenant_id, code, name, asset_type, unit_id, location, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, code, name, asset_type, unit_id, location, metadata, active, created_at;

-- name: ListMaintenancePlans :many
SELECT id, tenant_id, asset_id, name, frequency_days, frequency_km, frequency_hours,
       last_done, next_due, active
FROM erp_maintenance_plans WHERE tenant_id = $1 AND asset_id = $2 ORDER BY name;

-- name: CreateMaintenancePlan :one
INSERT INTO erp_maintenance_plans (tenant_id, asset_id, name, frequency_days, frequency_km, frequency_hours, next_due)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, asset_id, name, frequency_days, frequency_km, frequency_hours, last_done, next_due, active;

-- name: ListWorkOrders :many
SELECT wo.id, wo.tenant_id, wo.number, wo.asset_id, wo.date, wo.work_type,
       wo.description, wo.assigned_to, wo.status, wo.priority, wo.completed_at,
       wo.user_id, wo.notes, wo.created_at,
       a.code AS asset_code, a.name AS asset_name
FROM erp_work_orders wo
JOIN erp_maintenance_assets a ON a.id = wo.asset_id
WHERE wo.tenant_id = $1
  AND (sqlc.arg(status_filter)::TEXT = '' OR wo.status = sqlc.arg(status_filter)::TEXT)
ORDER BY wo.date DESC
LIMIT $2 OFFSET $3;

-- name: GetWorkOrder :one
SELECT id, tenant_id, number, asset_id, plan_id, date, work_type, description,
       assigned_to, status, priority, completed_at, user_id, notes, created_at
FROM erp_work_orders WHERE id = $1 AND tenant_id = $2;

-- name: CreateWorkOrder :one
INSERT INTO erp_work_orders (tenant_id, number, asset_id, plan_id, date, work_type,
    description, assigned_to, priority, user_id, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, tenant_id, number, asset_id, plan_id, date, work_type, description,
    assigned_to, status, priority, completed_at, user_id, notes, created_at;

-- name: UpdateWorkOrderStatus :execrows
UPDATE erp_work_orders SET status = $3,
    completed_at = CASE WHEN $3 = 'completed' THEN now() ELSE completed_at END
WHERE id = $1 AND tenant_id = $2;

-- name: ListWorkOrderParts :many
SELECT wop.id, wop.tenant_id, wop.work_order_id, wop.article_id, wop.quantity,
       a.code AS article_code, a.name AS article_name
FROM erp_work_order_parts wop
JOIN erp_articles a ON a.id = wop.article_id
WHERE wop.work_order_id = $1 AND wop.tenant_id = $2;

-- name: CreateWorkOrderPart :one
INSERT INTO erp_work_order_parts (tenant_id, work_order_id, article_id, quantity)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, work_order_id, article_id, quantity;

-- name: ListFuelLogs :many
SELECT fl.id, fl.tenant_id, fl.asset_id, fl.date, fl.liters, fl.km_reading,
       fl.cost, fl.user_id, fl.created_at,
       a.code AS asset_code, a.name AS asset_name
FROM erp_fuel_logs fl
JOIN erp_maintenance_assets a ON a.id = fl.asset_id
WHERE fl.tenant_id = $1
  AND (sqlc.arg(asset_filter)::UUID IS NULL OR fl.asset_id = sqlc.arg(asset_filter)::UUID)
ORDER BY fl.date DESC
LIMIT $2 OFFSET $3;

-- name: CreateFuelLog :one
INSERT INTO erp_fuel_logs (tenant_id, asset_id, date, liters, km_reading, cost, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, asset_id, date, liters, km_reading, cost, user_id, created_at;
