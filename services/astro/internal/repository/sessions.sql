-- name: CreateSession :one
INSERT INTO astro_sessions (tenant_id, user_id, contact_id, title)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, user_id, contact_id, title, pinned, created_at, updated_at;

-- name: GetSession :one
SELECT id, tenant_id, user_id, contact_id, title, pinned, created_at, updated_at
FROM astro_sessions WHERE tenant_id = $1 AND user_id = $2 AND id = $3;

-- name: ListSessions :many
SELECT id, tenant_id, user_id, contact_id, title, pinned, created_at, updated_at
FROM astro_sessions WHERE tenant_id = $1 AND user_id = $2
ORDER BY pinned DESC, updated_at DESC LIMIT $3 OFFSET $4;

-- name: UpdateSessionTitle :exec
UPDATE astro_sessions SET title = $4, updated_at = now()
WHERE tenant_id = $1 AND user_id = $2 AND id = $3;

-- name: UpdateSessionPinned :exec
UPDATE astro_sessions SET pinned = $4, updated_at = now()
WHERE tenant_id = $1 AND user_id = $2 AND id = $3;

-- name: DeleteSession :exec
DELETE FROM astro_sessions WHERE tenant_id = $1 AND user_id = $2 AND id = $3;

-- name: AddMessage :one
INSERT INTO astro_messages (tenant_id, session_id, role, content, thinking, techniques, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, session_id, role, content, thinking, techniques, metadata, created_at;

-- name: GetMessages :many
SELECT id, tenant_id, session_id, role, content, thinking, techniques, metadata, created_at
FROM astro_messages WHERE tenant_id = $1 AND session_id = $2
ORDER BY created_at ASC LIMIT $3;

-- name: TouchSession :exec
UPDATE astro_sessions SET updated_at = now()
WHERE id = $1;
