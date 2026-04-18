-- name: ListNonconformities :many
SELECT nc.id, nc.tenant_id, nc.number, nc.date, nc.description, nc.severity,
       nc.status, nc.assigned_to, nc.user_id, nc.created_at,
       e.name AS assigned_name
FROM erp_nonconformities nc
LEFT JOIN erp_entities e ON e.id = nc.assigned_to
WHERE nc.tenant_id = $1
  AND (sqlc.arg(status_filter)::TEXT = '' OR nc.status = sqlc.arg(status_filter)::TEXT)
  AND (sqlc.arg(severity_filter)::TEXT = '' OR nc.severity = sqlc.arg(severity_filter)::TEXT)
ORDER BY nc.date DESC
LIMIT $2 OFFSET $3;

-- name: GetNonconformity :one
SELECT id, tenant_id, number, date, type_id, origin_id, description, severity,
       status, assigned_to, closed_at, user_id, created_at
FROM erp_nonconformities WHERE id = $1 AND tenant_id = $2;

-- name: CreateNonconformity :one
INSERT INTO erp_nonconformities (tenant_id, number, date, type_id, origin_id, description, severity, assigned_to, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, tenant_id, number, date, type_id, origin_id, description, severity, status, assigned_to, closed_at, user_id, created_at;

-- name: UpdateNonconformityStatus :execrows
UPDATE erp_nonconformities SET status = $3,
    closed_at = CASE WHEN $3 = 'closed' THEN now() ELSE closed_at END
WHERE id = $1 AND tenant_id = $2;

-- name: ListCorrectiveActions :many
SELECT id, tenant_id, nc_id, action_type, description, responsible_id,
       due_date, status, completed_at, effectiveness
FROM erp_corrective_actions WHERE nc_id = $1 AND tenant_id = $2 ORDER BY due_date;

-- name: CreateCorrectiveAction :one
INSERT INTO erp_corrective_actions (tenant_id, nc_id, action_type, description, responsible_id, due_date)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tenant_id, nc_id, action_type, description, responsible_id, due_date, status, completed_at, effectiveness;

-- name: ListAudits :many
SELECT id, tenant_id, number, date, audit_type, scope, lead_auditor_id,
       status, score, notes, created_at
FROM erp_audits WHERE tenant_id = $1 ORDER BY date DESC LIMIT $2 OFFSET $3;

-- name: CreateAudit :one
INSERT INTO erp_audits (tenant_id, number, date, audit_type, scope, lead_auditor_id, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, number, date, audit_type, scope, lead_auditor_id, status, score, notes, created_at;

-- name: ListAuditFindings :many
SELECT id, tenant_id, audit_id, finding_type, description, nc_id
FROM erp_audit_findings WHERE audit_id = $1 AND tenant_id = $2;

-- name: CreateAuditFinding :one
INSERT INTO erp_audit_findings (tenant_id, audit_id, finding_type, description, nc_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, audit_id, finding_type, description, nc_id;

-- name: ListControlledDocuments :many
SELECT id, tenant_id, code, title, revision, doc_type_id, file_key,
       approved_by, approved_at, status, created_at
FROM erp_controlled_documents WHERE tenant_id = $1
  AND (sqlc.arg(status_filter)::TEXT = '' OR status = sqlc.arg(status_filter)::TEXT)
ORDER BY code, revision DESC
LIMIT $2 OFFSET $3;

-- name: CreateControlledDocument :one
INSERT INTO erp_controlled_documents (tenant_id, code, title, revision, doc_type_id, file_key)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tenant_id, code, title, revision, doc_type_id, file_key, approved_by, approved_at, status, created_at;

-- ─── Supplier Scorecards ───────────────────────────────────────────────────

-- name: ListSupplierScorecards :many
SELECT sc.*, e.name AS supplier_name
FROM erp_supplier_scorecards sc
JOIN erp_entities e ON e.id = sc.supplier_id AND e.tenant_id = sc.tenant_id
WHERE sc.tenant_id = $1
ORDER BY sc.period DESC, sc.quality_score ASC
LIMIT $2 OFFSET $3;

-- name: UpsertSupplierScorecard :one
INSERT INTO erp_supplier_scorecards (tenant_id, supplier_id, period, total_receipts, accepted_qty, rejected_qty, total_demerits, quality_score)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (tenant_id, supplier_id, period) DO UPDATE SET
    total_receipts = EXCLUDED.total_receipts, accepted_qty = EXCLUDED.accepted_qty,
    rejected_qty = EXCLUDED.rejected_qty, total_demerits = EXCLUDED.total_demerits,
    quality_score = EXCLUDED.quality_score
RETURNING *;

-- ─── Risk Register ─────────────────────────────────────────────────────────

-- name: ListQualityRisks :many
SELECT qr.*, e.name AS responsible_name
FROM erp_quality_risks qr
LEFT JOIN erp_entities e ON e.id = qr.responsible_id AND e.tenant_id = qr.tenant_id
WHERE qr.tenant_id = $1
    AND (sqlc.arg(status_filter)::TEXT = '' OR qr.status = sqlc.arg(status_filter)::TEXT)
ORDER BY
    CASE qr.probability WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 ELSE 4 END,
    CASE qr.impact WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 ELSE 4 END
LIMIT $2 OFFSET $3;

-- name: CreateQualityRisk :one
INSERT INTO erp_quality_risks (tenant_id, title, description, category, probability, impact, mitigation, responsible_id, review_date, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: UpdateQualityRiskStatus :one
UPDATE erp_quality_risks SET status = $3, updated_at = now() WHERE id = $1 AND tenant_id = $2 RETURNING *;

-- ─── Quality Indicators ────────────────────────────────────────────────────

-- name: ListQualityIndicators :many
SELECT * FROM erp_quality_indicators
WHERE tenant_id = $1
    AND period BETWEEN sqlc.arg(period_from) AND sqlc.arg(period_to)
ORDER BY period, indicator_type;

-- name: UpsertQualityIndicator :one
INSERT INTO erp_quality_indicators (tenant_id, period, indicator_type, value, target)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (tenant_id, period, indicator_type) DO UPDATE SET value = EXCLUDED.value, target = EXCLUDED.target
RETURNING *;

-- name: CalculateQualityKPIs :one
SELECT
    COUNT(*) FILTER (WHERE status != 'closed') AS open_ncs,
    COUNT(*) AS total_ncs,
    COUNT(*) FILTER (WHERE status = 'closed') AS closed_ncs,
    CASE WHEN COUNT(*) > 0 THEN
        ROUND(COUNT(*) FILTER (WHERE status = 'closed')::NUMERIC / COUNT(*)::NUMERIC * 100, 1)
    ELSE 0 END AS resolution_rate
FROM erp_nonconformities WHERE tenant_id = $1
    AND date >= sqlc.arg(date_from) AND date <= sqlc.arg(date_to);

-- ─── NC Origins ────────────────────────────────────────────────────────────

-- name: ListNCOrigins :many
SELECT * FROM erp_nc_origins WHERE tenant_id = $1 AND active = true ORDER BY name;

-- name: CreateNCOrigin :one
INSERT INTO erp_nc_origins (tenant_id, name) VALUES ($1, $2) RETURNING *;

-- ─── Quality Action Plans ───────────────────────────────────────────────────

-- name: ListActionPlans :many
SELECT ap.*, e.name AS responsible_name, d.name AS section_name,
       nc.number AS nc_number
FROM erp_quality_action_plans ap
LEFT JOIN erp_entities e ON e.id = ap.responsible_id AND e.tenant_id = ap.tenant_id
LEFT JOIN erp_departments d ON d.id = ap.section_id AND d.tenant_id = ap.tenant_id
LEFT JOIN erp_nonconformities nc ON nc.id = ap.nonconformity_id AND nc.tenant_id = ap.tenant_id
WHERE ap.tenant_id = $1
  AND (sqlc.arg(nc_filter)::TEXT = '' OR ap.nonconformity_id::TEXT = sqlc.arg(nc_filter)::TEXT)
  AND (sqlc.arg(status_filter)::TEXT = '' OR ap.status = sqlc.arg(status_filter)::TEXT)
ORDER BY ap.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetActionPlan :one
SELECT * FROM erp_quality_action_plans WHERE id = $1 AND tenant_id = $2;

-- name: CreateActionPlan :one
INSERT INTO erp_quality_action_plans (tenant_id, nonconformity_id, responsible_id, section_id, description, planned_start, target_date, time_savings_hours, cost_savings, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: UpdateActionPlanStatus :execrows
UPDATE erp_quality_action_plans
SET status = $3, closed_date = CASE WHEN $3 IN ('closed','cancelled') THEN now()::DATE ELSE closed_date END, updated_at = now()
WHERE id = $1 AND tenant_id = $2;

-- ─── Quality Action Tasks ───────────────────────────────────────────────────

-- name: ListActionTasks :many
SELECT at.*, e.name AS leader_name
FROM erp_quality_action_tasks at
LEFT JOIN erp_entities e ON e.id = at.leader_id AND e.tenant_id = at.tenant_id
WHERE at.tenant_id = $1 AND at.plan_id = $2
ORDER BY at.created_at ASC;

-- name: CreateActionTask :one
INSERT INTO erp_quality_action_tasks (tenant_id, plan_id, description, leader_id, planned_start, target_date)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: CompleteActionTask :execrows
UPDATE erp_quality_action_tasks
SET completed = true, closed_date = now()::DATE, updated_at = now()
WHERE id = $1 AND tenant_id = $2;

-- name: ListInspectionTemplates :many
-- Master list of inspection controls (migrated from Histrix PROD_CONTROLES).
-- Returns actives by default; section_id filter narrows to a specific
-- workshop area when used from the production-floor UI.
SELECT id, tenant_id, section_id, step_id, vehicle_section_id, control_name, model_code,
       control_type, sort_order, active, critical, actionable, show_in_tech_sheet,
       default_inspector_id, enabled_user_id, observations, metadata, created_at, updated_at
FROM erp_inspection_templates
WHERE tenant_id = $1
  AND (sqlc.arg(active_only)::BOOLEAN = false OR active = true)
  AND (sqlc.arg(section_filter)::UUID IS NULL OR section_id = sqlc.arg(section_filter)::UUID)
ORDER BY sort_order, control_name;
