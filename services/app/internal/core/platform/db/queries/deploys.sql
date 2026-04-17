-- name: InsertDeployLog :one
INSERT INTO deploy_log (service, version_from, version_to, status, deployed_by, notes)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, service, version_from, version_to, status, deployed_by, started_at, finished_at, notes;

-- name: ListDeployLogs :many
SELECT id, service, version_from, version_to, status, deployed_by, started_at, finished_at, notes
FROM deploy_log
ORDER BY started_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateDeployStatus :exec
UPDATE deploy_log
SET status = $2, finished_at = now()
WHERE id = $1;
