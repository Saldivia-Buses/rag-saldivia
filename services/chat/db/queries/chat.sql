-- name: CreateSession :one
INSERT INTO sessions (user_id, title, collection)
VALUES ($1, $2, $3)
RETURNING id, user_id, title, collection, is_saved, created_at, updated_at;

-- name: GetSession :one
SELECT id, user_id, title, collection, is_saved, created_at, updated_at
FROM sessions WHERE id = $1 AND user_id = $2;

-- name: ListSessionsByUser :many
SELECT id, user_id, title, collection, is_saved, created_at, updated_at
FROM sessions WHERE user_id = $1
ORDER BY updated_at DESC;

-- name: DeleteSession :execrows
DELETE FROM sessions WHERE id = $1 AND user_id = $2;

-- name: RenameSession :execrows
UPDATE sessions SET title = $3, updated_at = now()
WHERE id = $1 AND user_id = $2;

-- name: TouchSession :exec
UPDATE sessions SET updated_at = now() WHERE id = $1 AND user_id = $2;

-- name: CreateMessage :one
INSERT INTO messages (session_id, role, content, thinking, sources, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, session_id, role, content, thinking, sources, metadata, created_at;

-- name: ListMessages :many
SELECT id, session_id, role, content, thinking, sources, metadata, created_at
FROM messages WHERE session_id = $1
ORDER BY created_at;
