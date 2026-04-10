-- name: ListSuggestions :many
SELECT s.id, s.tenant_id, s.user_id, s.origin, s.body, s.is_read, s.created_at, s.updated_at,
       COUNT(r.id)::INT AS response_count
FROM erp_suggestions s
LEFT JOIN erp_suggestion_responses r ON r.suggestion_id = s.id AND r.tenant_id = s.tenant_id
WHERE s.tenant_id = $1
GROUP BY s.id
ORDER BY s.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetSuggestion :one
SELECT id, tenant_id, user_id, origin, body, is_read, created_at, updated_at
FROM erp_suggestions
WHERE id = $1 AND tenant_id = $2;

-- name: CreateSuggestion :one
INSERT INTO erp_suggestions (tenant_id, user_id, origin, body)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, user_id, origin, body, is_read, created_at, updated_at;

-- name: MarkSuggestionRead :exec
UPDATE erp_suggestions SET is_read = true, updated_at = now()
WHERE id = $1 AND tenant_id = $2;

-- name: ListResponses :many
SELECT id, tenant_id, suggestion_id, user_id, body, created_at
FROM erp_suggestion_responses
WHERE suggestion_id = $1 AND tenant_id = $2
ORDER BY created_at ASC;

-- name: CreateResponse :one
INSERT INTO erp_suggestion_responses (tenant_id, suggestion_id, user_id, body)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, suggestion_id, user_id, body, created_at;

-- name: CountUnread :one
SELECT COUNT(*)::INT AS count
FROM erp_suggestions
WHERE tenant_id = $1 AND is_read = false;
