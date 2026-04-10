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
    union_id, health_plan_id, schedule_type, category_id, encrypted_salary, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (entity_id) DO UPDATE SET
    department_id = EXCLUDED.department_id, position = EXCLUDED.position,
    hire_date = EXCLUDED.hire_date, union_id = EXCLUDED.union_id,
    health_plan_id = EXCLUDED.health_plan_id, schedule_type = EXCLUDED.schedule_type,
    category_id = EXCLUDED.category_id,
    metadata = EXCLUDED.metadata, updated_at = now()
RETURNING id, tenant_id, entity_id, department_id, position, hire_date,
    termination_date, union_id, health_plan_id, schedule_type,
    category_id, encrypted_salary, metadata, created_at, updated_at;

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
