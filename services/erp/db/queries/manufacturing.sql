-- ─── Chassis Brands ───────────────────────────────────────────────────────────

-- name: ListChassisBrands :many
SELECT id, tenant_id, code, name, active, created_at, updated_at
FROM erp_chassis_brands
WHERE tenant_id = $1 AND active = true
ORDER BY name;

-- name: CreateChassisBrand :one
INSERT INTO erp_chassis_brands (tenant_id, code, name)
VALUES ($1, $2, $3)
RETURNING id, tenant_id, code, name, active, created_at, updated_at;

-- ─── Chassis Models ────────────────────────────────────────────────────────────

-- name: ListChassisModels :many
SELECT cm.id, cm.tenant_id, cm.brand_id, cm.model_code, cm.description,
       cm.traction, cm.engine_location, cm.active, cm.created_at, cm.updated_at,
       cb.name AS brand_name
FROM erp_chassis_models cm
JOIN erp_chassis_brands cb ON cb.id = cm.brand_id AND cb.tenant_id = cm.tenant_id
WHERE cm.tenant_id = $1
  AND (sqlc.arg(brand_filter)::TEXT = '' OR cm.brand_id::TEXT = sqlc.arg(brand_filter)::TEXT)
  AND cm.active = true
ORDER BY cb.name, cm.description;

-- ─── Carroceria Models ─────────────────────────────────────────────────────────

-- name: ListCarroceriaModels :many
SELECT id, tenant_id, code, model_code, description, abbreviation, double_deck,
       axle_weight_pct, productive_hours_per_station, active, tech_sheet_image,
       created_at, updated_at
FROM erp_carroceria_models
WHERE tenant_id = $1 AND active = true
ORDER BY description;

-- name: GetCarroceriaModel :one
SELECT id, tenant_id, code, model_code, description, abbreviation, double_deck,
       axle_weight_pct, productive_hours_per_station, active, tech_sheet_image,
       created_at, updated_at
FROM erp_carroceria_models
WHERE id = $1 AND tenant_id = $2;

-- name: CreateCarroceriaModel :one
INSERT INTO erp_carroceria_models (
    tenant_id, code, model_code, description, abbreviation, double_deck,
    axle_weight_pct, productive_hours_per_station, tech_sheet_image
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, tenant_id, code, model_code, description, abbreviation, double_deck,
          axle_weight_pct, productive_hours_per_station, active, tech_sheet_image,
          created_at, updated_at;

-- ─── Carroceria BOM ────────────────────────────────────────────────────────────

-- name: GetCarroceriaBOM :many
SELECT bom.id, bom.tenant_id, bom.carroceria_model_id, bom.article_id,
       bom.quantity, bom.unit_of_use, bom.created_at, bom.updated_at,
       a.code AS article_code, a.name AS article_name
FROM erp_carroceria_bom bom
JOIN erp_articles a ON a.id = bom.article_id AND a.tenant_id = bom.tenant_id
WHERE bom.carroceria_model_id = $1 AND bom.tenant_id = $2
ORDER BY a.code;

-- name: AddBOMItem :one
INSERT INTO erp_carroceria_bom (tenant_id, carroceria_model_id, article_id, quantity, unit_of_use)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, carroceria_model_id, article_id, quantity, unit_of_use,
          created_at, updated_at;

-- name: DeleteBOMItem :execrows
DELETE FROM erp_carroceria_bom WHERE id = $1 AND tenant_id = $2;

-- ─── Manufacturing Units ───────────────────────────────────────────────────────

-- name: ListManufacturingUnits :many
SELECT mu.id, mu.tenant_id, mu.work_order_number, mu.chassis_serial, mu.engine_number,
       mu.chassis_brand_id, mu.chassis_model_id, mu.carroceria_model_id,
       mu.customer_id, mu.entry_date, mu.expected_completion, mu.actual_completion,
       mu.exit_date, mu.tachograph_id, mu.tachograph_serial, mu.invoice_reference,
       mu.observations, mu.status, mu.created_at, mu.updated_at,
       cb.name AS chassis_brand_name,
       cm.description AS chassis_model_description,
       crm.description AS carroceria_model_description,
       e.name AS customer_name
FROM erp_manufacturing_units mu
LEFT JOIN erp_chassis_brands cb ON cb.id = mu.chassis_brand_id AND cb.tenant_id = mu.tenant_id
LEFT JOIN erp_chassis_models cm ON cm.id = mu.chassis_model_id AND cm.tenant_id = mu.tenant_id
LEFT JOIN erp_carroceria_models crm ON crm.id = mu.carroceria_model_id AND crm.tenant_id = mu.tenant_id
LEFT JOIN erp_entities e ON e.id = mu.customer_id AND e.tenant_id = mu.tenant_id
WHERE mu.tenant_id = $1
  AND (sqlc.arg(status_filter)::TEXT = '' OR mu.status = sqlc.arg(status_filter)::TEXT)
ORDER BY mu.work_order_number DESC
LIMIT $2 OFFSET $3;

-- name: GetManufacturingUnit :one
SELECT mu.id, mu.tenant_id, mu.work_order_number, mu.chassis_serial, mu.engine_number,
       mu.chassis_brand_id, mu.chassis_model_id, mu.carroceria_model_id,
       mu.customer_id, mu.entry_date, mu.expected_completion, mu.actual_completion,
       mu.exit_date, mu.tachograph_id, mu.tachograph_serial, mu.invoice_reference,
       mu.observations, mu.status, mu.created_at, mu.updated_at,
       cb.name AS chassis_brand_name,
       cm.description AS chassis_model_description,
       crm.description AS carroceria_model_description,
       e.name AS customer_name
FROM erp_manufacturing_units mu
LEFT JOIN erp_chassis_brands cb ON cb.id = mu.chassis_brand_id AND cb.tenant_id = mu.tenant_id
LEFT JOIN erp_chassis_models cm ON cm.id = mu.chassis_model_id AND cm.tenant_id = mu.tenant_id
LEFT JOIN erp_carroceria_models crm ON crm.id = mu.carroceria_model_id AND crm.tenant_id = mu.tenant_id
LEFT JOIN erp_entities e ON e.id = mu.customer_id AND e.tenant_id = mu.tenant_id
WHERE mu.id = $1 AND mu.tenant_id = $2;

-- name: CreateManufacturingUnit :one
INSERT INTO erp_manufacturing_units (
    tenant_id, work_order_number, chassis_serial, engine_number,
    chassis_brand_id, chassis_model_id, carroceria_model_id, customer_id,
    entry_date, expected_completion, tachograph_id, tachograph_serial,
    invoice_reference, observations
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING id, tenant_id, work_order_number, chassis_serial, engine_number,
          chassis_brand_id, chassis_model_id, carroceria_model_id, customer_id,
          entry_date, expected_completion, actual_completion, exit_date,
          tachograph_id, tachograph_serial, invoice_reference, observations,
          status, created_at, updated_at;

-- name: UpdateManufacturingUnitStatus :execrows
UPDATE erp_manufacturing_units
SET status = $3, updated_at = now()
WHERE id = $1 AND tenant_id = $2;

-- name: UpdateUnitCompletion :execrows
UPDATE erp_manufacturing_units
SET actual_completion = $3, exit_date = $4, updated_at = now()
WHERE id = $1 AND tenant_id = $2;

-- ─── Production Control Causals ────────────────────────────────────────────────

-- name: ListControlCausals :many
SELECT id, tenant_id, code, description, causal_type, active, created_at, updated_at
FROM erp_production_control_causals
WHERE tenant_id = $1 AND active = true
ORDER BY causal_type, code;

-- ─── Production Controls (station tracking per unit) ──────────────────────────

-- name: ListProductionControls :many
SELECT pc.id, pc.tenant_id, pc.unit_id, pc.station, pc.station_seq,
       pc.responsible_id, pc.planned_start, pc.planned_end, pc.actual_start,
       pc.actual_end, pc.status, pc.notes, pc.created_at, pc.updated_at,
       e.name AS responsible_name
FROM erp_production_controls pc
LEFT JOIN erp_entities e ON e.id = pc.responsible_id AND e.tenant_id = pc.tenant_id
WHERE pc.tenant_id = $1 AND pc.unit_id = $2
ORDER BY pc.station_seq;

-- name: GetProductionControl :one
SELECT id, tenant_id, unit_id, station, station_seq, responsible_id,
       planned_start, planned_end, actual_start, actual_end, status, notes,
       created_at, updated_at
FROM erp_production_controls
WHERE id = $1 AND tenant_id = $2;

-- name: CreateProductionControl :one
INSERT INTO erp_production_controls (
    tenant_id, unit_id, station, station_seq, responsible_id,
    planned_start, planned_end, notes
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, tenant_id, unit_id, station, station_seq, responsible_id,
          planned_start, planned_end, actual_start, actual_end, status, notes,
          created_at, updated_at;

-- ─── Control Executions (time entries per station) ────────────────────────────

-- name: GetUnitControlExecutions :many
SELECT pce.id, pce.tenant_id, pce.control_id, pce.operator_id, pce.causal_id,
       pce.started_at, pce.ended_at, pce.duration, pce.exec_type, pce.notes,
       pce.created_at,
       pc.station AS station_name,
       caus.code AS causal_code, caus.description AS causal_description,
       e.name AS operator_name
FROM erp_production_control_executions pce
JOIN erp_production_controls pc ON pc.id = pce.control_id AND pc.tenant_id = pce.tenant_id
LEFT JOIN erp_production_control_causals caus ON caus.id = pce.causal_id AND caus.tenant_id = pce.tenant_id
LEFT JOIN erp_entities e ON e.id = pce.operator_id AND e.tenant_id = pce.tenant_id
WHERE pc.unit_id = $1 AND pce.tenant_id = $2
ORDER BY pce.started_at DESC;

-- name: GetUnitPendingControls :many
SELECT pc.id, pc.tenant_id, pc.unit_id, pc.station, pc.station_seq,
       pc.responsible_id, pc.planned_start, pc.planned_end, pc.actual_start,
       pc.actual_end, pc.status, pc.notes, pc.created_at, pc.updated_at
FROM erp_production_controls pc
WHERE pc.tenant_id = $1
  AND pc.unit_id = sqlc.arg(unit_id)::UUID
  AND pc.status IN ('pending', 'in_progress', 'blocked', 'rework')
ORDER BY pc.station_seq;

-- name: ExecuteControl :one
INSERT INTO erp_production_control_executions (
    tenant_id, control_id, operator_id, causal_id, started_at, exec_type, notes
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, control_id, operator_id, causal_id, started_at,
          ended_at, duration, exec_type, notes, created_at;

-- name: GetControlExecution :one
SELECT id, tenant_id, control_id, operator_id, causal_id, started_at,
       ended_at, duration, exec_type, notes, created_at
FROM erp_production_control_executions
WHERE id = $1 AND tenant_id = $2;

-- ─── Production Rework ─────────────────────────────────────────────────────────

-- name: CreateRework :one
INSERT INTO erp_production_rework (
    tenant_id, control_id, causal_id, reported_by, defect_desc, severity
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tenant_id, control_id, causal_id, reported_by, corrected_by,
          defect_desc, correction, severity, reported_at, resolved_at,
          created_at, updated_at;

-- name: ApproveRework :execrows
UPDATE erp_production_rework
SET corrected_by = $3, correction = $4, resolved_at = now(), updated_at = now()
WHERE id = $1 AND tenant_id = $2;

-- ─── LCM (Material Consumption per Unit) ─────────────────────────────────────

-- name: ListLCM :many
SELECT lcm.id, lcm.tenant_id, lcm.unit_id, lcm.issued_at, lcm.issued_by,
       lcm.warehouse_id, lcm.reference, lcm.status, lcm.notes,
       lcm.created_at, lcm.updated_at,
       e.name AS issued_by_name
FROM erp_manufacturing_lcm lcm
LEFT JOIN erp_entities e ON e.id = lcm.issued_by AND e.tenant_id = lcm.tenant_id
WHERE lcm.tenant_id = $1
  AND (sqlc.arg(unit_filter)::TEXT = '' OR lcm.unit_id::TEXT = sqlc.arg(unit_filter)::TEXT)
  AND (sqlc.arg(status_filter)::TEXT = '' OR lcm.status = sqlc.arg(status_filter)::TEXT)
ORDER BY lcm.issued_at DESC
LIMIT $2 OFFSET $3;

-- name: GetLCM :one
SELECT id, tenant_id, unit_id, issued_at, issued_by, warehouse_id,
       reference, status, notes, created_at, updated_at
FROM erp_manufacturing_lcm
WHERE id = $1 AND tenant_id = $2;

-- name: CreateLCM :one
INSERT INTO erp_manufacturing_lcm (
    tenant_id, unit_id, issued_by, warehouse_id, reference, notes
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tenant_id, unit_id, issued_at, issued_by, warehouse_id,
          reference, status, notes, created_at, updated_at;

-- ─── LCM Line Items (articles per LCM) ───────────────────────────────────────

-- name: ListLCMModels :many
SELECT lm.id, lm.tenant_id, lm.lcm_id, lm.article_id, lm.bom_qty,
       lm.issued_qty, lm.returned_qty, lm.unit_cost, lm.notes,
       lm.created_at, lm.updated_at,
       a.code AS article_code, a.name AS article_name
FROM erp_manufacturing_lcm_models lm
JOIN erp_articles a ON a.id = lm.article_id AND a.tenant_id = lm.tenant_id
WHERE lm.lcm_id = $1 AND lm.tenant_id = $2
ORDER BY a.code;

-- name: AddLCMModel :one
INSERT INTO erp_manufacturing_lcm_models (
    tenant_id, lcm_id, article_id, bom_qty, issued_qty, returned_qty, unit_cost, notes
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, tenant_id, lcm_id, article_id, bom_qty, issued_qty, returned_qty,
          unit_cost, notes, created_at, updated_at;

-- ─── CNRT Work Orders ─────────────────────────────────────────────────────────

-- name: ListCNRTWork :many
SELECT id, tenant_id, unit_id, cnrt_number, inspection_type, inspector_name,
       inspection_date, approved, approval_date, expiry_date, observations,
       rejection_reasons, status, document_url, created_at, updated_at
FROM erp_cnrt_work_orders
WHERE tenant_id = $1 AND unit_id = $2
ORDER BY inspection_date DESC;

-- name: CreateCNRTWork :one
INSERT INTO erp_cnrt_work_orders (
    tenant_id, unit_id, cnrt_number, inspection_type, inspector_name,
    inspection_date, observations
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, unit_id, cnrt_number, inspection_type, inspector_name,
          inspection_date, approved, approval_date, expiry_date, observations,
          rejection_reasons, status, document_url, created_at, updated_at;

-- ─── Manufacturing Certificates ───────────────────────────────────────────────

-- name: GetCertificate :one
SELECT id, tenant_id, unit_id, certificate_number, cert_type, issued_by,
       issued_at, valid_from, valid_until, authority, document_url,
       observations, status, revoked_at, revocation_reason, created_at, updated_at
FROM erp_manufacturing_certificates
WHERE tenant_id = $1 AND unit_id = $2
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateCertificate :one
INSERT INTO erp_manufacturing_certificates (
    tenant_id, unit_id, certificate_number, cert_type, issued_by,
    valid_from, valid_until, authority, document_url, observations
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, tenant_id, unit_id, certificate_number, cert_type, issued_by,
          issued_at, valid_from, valid_until, authority, document_url,
          observations, status, revoked_at, revocation_reason, created_at, updated_at;

-- name: UpdateCertificate :one
UPDATE erp_manufacturing_certificates SET
    certificate_number = $3,
    cert_type = $4,
    valid_from = $5,
    valid_until = $6,
    authority = $7,
    document_url = $8,
    observations = $9,
    updated_at = now()
WHERE id = $1 AND tenant_id = $2
RETURNING id, tenant_id, unit_id, certificate_number, cert_type, issued_by,
          issued_at, valid_from, valid_until, authority, document_url,
          observations, status, revoked_at, revocation_reason, created_at, updated_at;

-- name: IssueCertificate :execrows
UPDATE erp_manufacturing_certificates
SET status = 'issued', issued_at = now(), updated_at = now()
WHERE id = $1 AND tenant_id = $2;

-- ─── Analytics ─────────────────────────────────────────────────────────────────

-- name: GetManufacturingKPIs :one
SELECT
    COUNT(*) AS total_units,
    COUNT(*) FILTER (WHERE status = 'pending') AS pending,
    COUNT(*) FILTER (WHERE status = 'in_production') AS in_production,
    COUNT(*) FILTER (WHERE status = 'completed') AS completed,
    COUNT(*) FILTER (WHERE status = 'delivered') AS delivered,
    COUNT(*) FILTER (WHERE status = 'returned') AS returned,
    COUNT(*) FILTER (WHERE expected_completion < CURRENT_DATE AND actual_completion IS NULL) AS overdue
FROM erp_manufacturing_units
WHERE tenant_id = $1;
