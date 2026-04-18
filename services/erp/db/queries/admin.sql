-- name: ListCommunications :many
SELECT c.id, c.tenant_id, c.subject, c.body, c.sender_id, c.priority, c.created_at
FROM erp_communications c WHERE c.tenant_id = $1
ORDER BY c.created_at DESC LIMIT $2 OFFSET $3;

-- name: CreateCommunication :one
INSERT INTO erp_communications (tenant_id, subject, body, sender_id, priority)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, subject, body, sender_id, priority, created_at;

-- name: ListCalendarEvents :many
SELECT id, tenant_id, title, description, start_at, end_at, all_day, entity_id, user_id, created_at
FROM erp_calendar_events WHERE tenant_id = $1
  AND (sqlc.arg(date_from)::TIMESTAMPTZ IS NULL OR start_at >= sqlc.arg(date_from)::TIMESTAMPTZ)
  AND (sqlc.arg(date_to)::TIMESTAMPTZ IS NULL OR start_at <= sqlc.arg(date_to)::TIMESTAMPTZ)
ORDER BY start_at;

-- name: CreateCalendarEvent :one
INSERT INTO erp_calendar_events (tenant_id, title, description, start_at, end_at, all_day, entity_id, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, tenant_id, title, description, start_at, end_at, all_day, entity_id, user_id, created_at;

-- name: ListSurveys :many
SELECT id, tenant_id, title, description, status, user_id, created_at
FROM erp_surveys WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: CreateSurvey :one
INSERT INTO erp_surveys (tenant_id, title, description, user_id)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, title, description, status, user_id, created_at;

-- name: ListLegacyArchive :many
-- Paginated admin read of the legacy-archive JSONB store. Populated by the
-- migrate-legacy CLI when --archive-skips / --archive is set. Filter by
-- legacy_table so the UI can drill into a specific source table's skipped
-- rows without loading the full 7M-row archive.
SELECT id, tenant_id, legacy_table, legacy_pk, legacy_pk_num, data, migrated_at
FROM erp_legacy_archive
WHERE tenant_id = $1
  AND (sqlc.arg(legacy_table_filter)::TEXT = '' OR legacy_table = sqlc.arg(legacy_table_filter)::TEXT)
ORDER BY migrated_at DESC
LIMIT $2 OFFSET $3;

-- name: CountLegacyArchiveByTable :many
-- Summary for the admin UI: how many rows per legacy_table landed in the
-- archive. Cheap enough to run on demand (indexed on (tenant_id, legacy_table)).
SELECT legacy_table, COUNT(*) AS rows
FROM erp_legacy_archive WHERE tenant_id = $1
GROUP BY legacy_table
ORDER BY rows DESC;
