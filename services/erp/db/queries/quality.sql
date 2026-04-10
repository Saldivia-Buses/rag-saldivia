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
