-- name: ListDepartments :many
SELECT id, tenant_id, code, name, parent_id, manager_id, active
FROM erp_departments WHERE tenant_id = $1 ORDER BY code;

-- name: CreateDepartment :one
INSERT INTO erp_departments (tenant_id, code, name, parent_id, manager_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, code, name, parent_id, manager_id, active;

-- name: ListEmployeeDetails :many
SELECT ed.id, ed.tenant_id, ed.entity_id, ed.department_id, ed.position,
       ed.hire_date, ed.termination_date, ed.schedule_type, ed.created_at,
       e.name AS entity_name, e.code AS entity_code
FROM erp_employee_details ed
JOIN erp_entities e ON e.id = ed.entity_id
WHERE ed.tenant_id = $1
ORDER BY e.name
LIMIT $2 OFFSET $3;

-- name: GetEmployeeDetail :one
SELECT id, tenant_id, entity_id, department_id, position, hire_date,
       termination_date, union_id, health_plan_id, schedule_type,
       category_id, metadata, created_at, updated_at
FROM erp_employee_details WHERE entity_id = $1 AND tenant_id = $2;

-- name: UpsertEmployeeDetail :one
INSERT INTO erp_employee_details (tenant_id, entity_id, department_id, position, hire_date,
    union_id, health_plan_id, schedule_type, category_id, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (entity_id) DO UPDATE SET
    department_id = EXCLUDED.department_id, position = EXCLUDED.position,
    hire_date = EXCLUDED.hire_date, union_id = EXCLUDED.union_id,
    health_plan_id = EXCLUDED.health_plan_id, schedule_type = EXCLUDED.schedule_type,
    category_id = EXCLUDED.category_id,
    metadata = EXCLUDED.metadata, updated_at = now()
RETURNING id, tenant_id, entity_id, department_id, position, hire_date,
    termination_date, union_id, health_plan_id, schedule_type,
    category_id, metadata, created_at, updated_at;

-- name: ListHREvents :many
SELECT id, tenant_id, entity_id, event_type, date_from, date_to, hours,
       reason_id, notes, user_id, created_at
FROM erp_hr_events WHERE tenant_id = $1
  AND (sqlc.arg(entity_filter)::UUID IS NULL OR entity_id = sqlc.arg(entity_filter)::UUID)
  AND (sqlc.arg(type_filter)::TEXT = '' OR event_type = sqlc.arg(type_filter)::TEXT)
ORDER BY date_from DESC
LIMIT $2 OFFSET $3;

-- name: CreateHREvent :one
INSERT INTO erp_hr_events (tenant_id, entity_id, event_type, date_from, date_to, hours, reason_id, notes, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, tenant_id, entity_id, event_type, date_from, date_to, hours, reason_id, notes, user_id, created_at;

-- name: ListTraining :many
SELECT id, tenant_id, name, description, instructor, date_from, date_to, status
FROM erp_training WHERE tenant_id = $1 ORDER BY date_from DESC LIMIT $2 OFFSET $3;

-- name: CreateTraining :one
INSERT INTO erp_training (tenant_id, name, description, instructor, date_from, date_to)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tenant_id, name, description, instructor, date_from, date_to, status;

-- name: ListAttendance :many
SELECT id, tenant_id, entity_id, date, clock_in, clock_out, hours, source, created_at
FROM erp_attendance WHERE tenant_id = $1
  AND (sqlc.arg(entity_filter)::UUID IS NULL OR entity_id = sqlc.arg(entity_filter)::UUID)
  AND (sqlc.arg(date_from)::DATE IS NULL OR date >= sqlc.arg(date_from)::DATE)
  AND (sqlc.arg(date_to)::DATE IS NULL OR date <= sqlc.arg(date_to)::DATE)
ORDER BY date DESC, created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateAttendance :one
INSERT INTO erp_attendance (tenant_id, entity_id, date, clock_in, clock_out, hours, source)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, entity_id, date, clock_in, clock_out, hours, source, created_at;

-- ─── Competencies ──────────────────────────────────────────────────────────

-- name: ListCompetencies :many
SELECT * FROM erp_competencies WHERE tenant_id = $1 AND (sqlc.arg(active_only)::BOOLEAN = false OR active = true) ORDER BY category, name;

-- name: CreateCompetency :one
INSERT INTO erp_competencies (tenant_id, name, description, category) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: ListEmployeeCompetencies :many
SELECT ec.*, c.name AS competency_name, c.category AS competency_category
FROM erp_employee_competencies ec
JOIN erp_competencies c ON c.id = ec.competency_id AND c.tenant_id = ec.tenant_id
WHERE ec.tenant_id = $1 AND ec.entity_id = $2
ORDER BY c.category, c.name;

-- name: UpsertEmployeeCompetency :one
INSERT INTO erp_employee_competencies (tenant_id, entity_id, competency_id, level, certified, certified_at, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (tenant_id, entity_id, competency_id) DO UPDATE SET
    level = EXCLUDED.level, certified = EXCLUDED.certified,
    certified_at = EXCLUDED.certified_at, notes = EXCLUDED.notes, updated_at = now()
RETURNING *;

-- ─── Evaluations ───────────────────────────────────────────────────────────

-- name: ListEvaluations :many
SELECT ev.*, e.name AS entity_name
FROM erp_evaluations ev
JOIN erp_entities e ON e.id = ev.entity_id AND e.tenant_id = ev.tenant_id
WHERE ev.tenant_id = $1
    AND (sqlc.arg(period_filter)::TEXT = '' OR ev.period = sqlc.arg(period_filter)::TEXT)
ORDER BY ev.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetEvaluation :one
SELECT ev.*, e.name AS entity_name
FROM erp_evaluations ev
JOIN erp_entities e ON e.id = ev.entity_id AND e.tenant_id = ev.tenant_id
WHERE ev.id = $1 AND ev.tenant_id = $2;

-- name: CreateEvaluation :one
INSERT INTO erp_evaluations (tenant_id, entity_id, evaluator_id, period, eval_type, strengths, weaknesses, goals, comments)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: SubmitEvaluation :one
UPDATE erp_evaluations SET overall_score = $3, status = 'submitted', submitted_at = now()
WHERE id = $1 AND tenant_id = $2 AND status = 'draft'
RETURNING *;

-- name: ListEvaluationScores :many
SELECT es.*, c.name AS competency_name
FROM erp_evaluation_scores es
JOIN erp_competencies c ON c.id = es.competency_id AND c.tenant_id = es.tenant_id
WHERE es.evaluation_id = $1 AND es.tenant_id = $2
ORDER BY c.category, c.name;

-- name: CreateEvaluationScore :one
INSERT INTO erp_evaluation_scores (tenant_id, evaluation_id, competency_id, score, comments)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- ─── Leave Balances ────────────────────────────────────────────────────────

-- name: ListLeaveBalances :many
SELECT * FROM erp_leave_balances
WHERE tenant_id = $1 AND entity_id = $2
ORDER BY year DESC, leave_type;

-- name: UpsertLeaveBalance :one
INSERT INTO erp_leave_balances (tenant_id, entity_id, leave_type, year, accrued, used)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (tenant_id, entity_id, leave_type, year) DO UPDATE SET
    accrued = EXCLUDED.accrued, used = EXCLUDED.used
RETURNING *;

-- name: GetSeniorityYears :one
SELECT
    EXTRACT(YEAR FROM AGE(CURRENT_DATE, ed.hire_date))::INT AS seniority_years,
    ed.hire_date
FROM erp_employee_details ed
WHERE ed.tenant_id = $1 AND ed.entity_id = $2 AND ed.hire_date IS NOT NULL;
