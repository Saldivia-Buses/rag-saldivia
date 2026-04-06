-- name: CreateJob :one
INSERT INTO ingest_jobs (user_id, collection, file_name, file_size, status)
VALUES ($1, $2, $3, $4, 'pending')
RETURNING id, user_id, collection, file_name, file_size, status, error, created_at, updated_at;

-- name: GetJob :one
SELECT id, user_id, collection, file_name, file_size, status, error, created_at, updated_at
FROM ingest_jobs WHERE id = $1 AND user_id = $2;

-- name: ListJobsByUser :many
SELECT id, user_id, collection, file_name, file_size, status, error, created_at, updated_at
FROM ingest_jobs WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2;

-- name: DeleteJob :execrows
DELETE FROM ingest_jobs WHERE id = $1 AND user_id = $2;

-- name: DeleteJobByID :exec
DELETE FROM ingest_jobs WHERE id = $1;

-- name: UpdateJobStatus :exec
UPDATE ingest_jobs SET status = $1, updated_at = now() WHERE id = $2;

-- name: UpdateJobStatusWithError :exec
UPDATE ingest_jobs SET status = $1, error = $2, updated_at = now() WHERE id = $3;
