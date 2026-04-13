-- ─── Accident Types ────────────────────────────────────────────────────────

-- name: ListAccidentTypes :many
SELECT id, tenant_id, name, abbreviation, severity_idx, active, created_at
FROM erp_accident_types
WHERE tenant_id = $1 AND active = true
ORDER BY name;

-- name: CreateAccidentType :one
INSERT INTO erp_accident_types (tenant_id, name, abbreviation, severity_idx)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, name, abbreviation, severity_idx, active, created_at;

-- ─── Body Parts ─────────────────────────────────────────────────────────────

-- name: ListBodyParts :many
SELECT id, tenant_id, description, created_at
FROM erp_body_parts
WHERE tenant_id = $1
ORDER BY description;

-- name: CreateBodyPart :one
INSERT INTO erp_body_parts (tenant_id, description)
VALUES ($1, $2)
RETURNING id, tenant_id, description, created_at;

-- ─── Risk Agents ─────────────────────────────────────────────────────────────

-- name: ListRiskAgents :many
SELECT id, tenant_id, name, risk_type, active, created_at
FROM erp_risk_agents
WHERE tenant_id = $1
  AND (sqlc.arg(risk_type_filter)::TEXT = '' OR risk_type = sqlc.arg(risk_type_filter)::TEXT)
ORDER BY name;

-- name: CreateRiskAgent :one
INSERT INTO erp_risk_agents (tenant_id, name, risk_type)
VALUES ($1, $2, $3)
RETURNING id, tenant_id, name, risk_type, active, created_at;

-- ─── Work Accidents ──────────────────────────────────────────────────────────

-- name: ListWorkAccidents :many
SELECT wa.id, wa.tenant_id, wa.entity_id, wa.accident_type_id, wa.body_part_id,
       wa.section_id, wa.incident_date, wa.recovery_date, wa.lost_days,
       wa.observations, wa.reported_by, wa.status, wa.created_at, wa.updated_at,
       e.name AS entity_name,
       at.name AS accident_type_name,
       bp.description AS body_part_description
FROM erp_work_accidents wa
LEFT JOIN erp_entities e ON e.id = wa.entity_id AND e.tenant_id = wa.tenant_id
LEFT JOIN erp_accident_types at ON at.id = wa.accident_type_id AND at.tenant_id = wa.tenant_id
LEFT JOIN erp_body_parts bp ON bp.id = wa.body_part_id AND bp.tenant_id = wa.tenant_id
WHERE wa.tenant_id = $1
  AND (sqlc.arg(status_filter)::TEXT = '' OR wa.status = sqlc.arg(status_filter)::TEXT)
ORDER BY wa.incident_date DESC
LIMIT $2 OFFSET $3;

-- name: GetWorkAccident :one
SELECT id, tenant_id, entity_id, accident_type_id, body_part_id, section_id,
       incident_date, recovery_date, lost_days, observations, reported_by,
       status, created_at, updated_at
FROM erp_work_accidents
WHERE id = $1 AND tenant_id = $2;

-- name: CreateWorkAccident :one
INSERT INTO erp_work_accidents (tenant_id, entity_id, accident_type_id, body_part_id,
    section_id, incident_date, recovery_date, lost_days, observations, reported_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, tenant_id, entity_id, accident_type_id, body_part_id, section_id,
          incident_date, recovery_date, lost_days, observations, reported_by,
          status, created_at, updated_at;

-- name: UpdateAccidentStatus :execrows
UPDATE erp_work_accidents SET status = $3, updated_at = now()
WHERE id = $1 AND tenant_id = $2;

-- ─── Risk Exposures ──────────────────────────────────────────────────────────

-- name: ListEmployeeRiskExposures :many
SELECT er.id, er.tenant_id, er.entity_id, er.risk_agent_id, er.section_id,
       er.exposed_from, er.exposed_until, er.notes, er.created_at, er.updated_at,
       ra.name AS risk_agent_name, ra.risk_type,
       e.name AS entity_name
FROM erp_employee_risk_exposures er
LEFT JOIN erp_risk_agents ra ON ra.id = er.risk_agent_id AND ra.tenant_id = er.tenant_id
LEFT JOIN erp_entities e ON e.id = er.entity_id AND e.tenant_id = er.tenant_id
WHERE er.tenant_id = $1
  AND (sqlc.arg(entity_filter)::TEXT = '' OR er.entity_id::TEXT = sqlc.arg(entity_filter)::TEXT)
ORDER BY er.exposed_from DESC;

-- name: CreateRiskExposure :one
INSERT INTO erp_employee_risk_exposures (tenant_id, entity_id, risk_agent_id, section_id,
    exposed_from, exposed_until, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, entity_id, risk_agent_id, section_id, exposed_from,
          exposed_until, notes, created_at, updated_at;

-- ─── Medical Consultations ───────────────────────────────────────────────────

-- name: ListMedicalConsultations :many
SELECT id, tenant_id, entity_id, patient_name, consult_date, consult_time,
       symptoms, prescription, medic_user, created_at
FROM erp_medical_consultations
WHERE tenant_id = $1
  AND (sqlc.arg(date_from)::DATE = '0001-01-01' OR consult_date >= sqlc.arg(date_from)::DATE)
  AND (sqlc.arg(date_to)::DATE = '0001-01-01' OR consult_date <= sqlc.arg(date_to)::DATE)
ORDER BY consult_date DESC
LIMIT $2 OFFSET $3;

-- name: CreateMedicalConsultation :one
INSERT INTO erp_medical_consultations (tenant_id, entity_id, patient_name, consult_date,
    consult_time, symptoms, prescription, medic_user)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, tenant_id, entity_id, patient_name, consult_date, consult_time,
          symptoms, prescription, medic_user, created_at;

-- ─── Medical Leaves ──────────────────────────────────────────────────────────

-- name: ListMedicalLeaves :many
SELECT ml.id, ml.tenant_id, ml.entity_id, ml.body_part_id, ml.accident_id,
       ml.leave_type, ml.date_from, ml.date_to, ml.working_days, ml.observations,
       ml.status, ml.approved_by, ml.approved_at, ml.created_at, ml.updated_at,
       e.name AS entity_name,
       bp.description AS body_part_description
FROM erp_medical_leaves ml
LEFT JOIN erp_entities e ON e.id = ml.entity_id AND e.tenant_id = ml.tenant_id
LEFT JOIN erp_body_parts bp ON bp.id = ml.body_part_id AND bp.tenant_id = ml.tenant_id
WHERE ml.tenant_id = $1
  AND (sqlc.arg(entity_filter)::TEXT = '' OR ml.entity_id::TEXT = sqlc.arg(entity_filter)::TEXT)
  AND (sqlc.arg(leave_type_filter)::TEXT = '' OR ml.leave_type = sqlc.arg(leave_type_filter)::TEXT)
  AND (sqlc.arg(status_filter)::TEXT = '' OR ml.status = sqlc.arg(status_filter)::TEXT)
ORDER BY ml.date_from DESC
LIMIT $2 OFFSET $3;

-- name: CreateMedicalLeave :one
INSERT INTO erp_medical_leaves (tenant_id, entity_id, body_part_id, accident_id,
    leave_type, date_from, date_to, working_days, observations)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, tenant_id, entity_id, body_part_id, accident_id, leave_type,
          date_from, date_to, working_days, observations, status,
          approved_by, approved_at, created_at, updated_at;

-- name: ApproveMedicalLeave :execrows
UPDATE erp_medical_leaves
SET status = 'approved', approved_by = $3, approved_at = now(), updated_at = now()
WHERE id = $1 AND tenant_id = $2;

-- ─── Safety KPIs ─────────────────────────────────────────────────────────────

-- name: GetSafetyKPIs :one
SELECT
    COUNT(*)                                      AS total_accidents,
    COUNT(*) FILTER (WHERE status = 'open')       AS open_accidents,
    COUNT(*) FILTER (WHERE status = 'closed')     AS closed_accidents,
    COALESCE(SUM(lost_days), 0)                   AS total_lost_days
FROM erp_work_accidents
WHERE tenant_id = $1
  AND incident_date >= sqlc.arg(date_from)::DATE
  AND incident_date <= sqlc.arg(date_to)::DATE;
