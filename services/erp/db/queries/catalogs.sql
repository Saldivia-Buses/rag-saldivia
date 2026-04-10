-- name: ListCatalogs :many
SELECT id, tenant_id, type, code, name, parent_id, sort_order, active, metadata, created_at, updated_at
FROM erp_catalogs
WHERE tenant_id = $1 AND type = $2 AND (sqlc.arg(active_only)::BOOLEAN = false OR active = true)
ORDER BY sort_order, name;

-- name: ListCatalogTypes :many
SELECT DISTINCT type FROM erp_catalogs
WHERE tenant_id = $1
ORDER BY type;

-- name: GetCatalog :one
SELECT id, tenant_id, type, code, name, parent_id, sort_order, active, metadata, created_at, updated_at
FROM erp_catalogs
WHERE id = $1 AND tenant_id = $2;

-- name: CreateCatalog :one
INSERT INTO erp_catalogs (tenant_id, type, code, name, parent_id, sort_order, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, type, code, name, parent_id, sort_order, active, metadata, created_at, updated_at;

-- name: UpdateCatalog :one
UPDATE erp_catalogs
SET code = $3, name = $4, parent_id = $5, sort_order = $6, active = $7, metadata = $8, updated_at = now()
WHERE id = $1 AND tenant_id = $2
RETURNING id, tenant_id, type, code, name, parent_id, sort_order, active, metadata, created_at, updated_at;

-- name: DeleteCatalog :execrows
UPDATE erp_catalogs SET active = false, updated_at = now()
WHERE id = $1 AND tenant_id = $2;
