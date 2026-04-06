-- name: GetTenantBySlug :one
SELECT id, slug, name, plan_id, postgres_url, redis_url, enabled, logo_url, domain, settings, created_at, updated_at
FROM tenants
WHERE slug = $1 AND enabled = true;

-- name: ListTenants :many
SELECT id, slug, name, plan_id, enabled, created_at
FROM tenants
ORDER BY created_at
LIMIT $1 OFFSET $2;

-- name: CreateTenant :one
INSERT INTO tenants (slug, name, plan_id, postgres_url, redis_url, settings)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, slug, name, plan_id, enabled, logo_url, domain, settings, created_at, updated_at;

-- name: UpdateTenant :exec
UPDATE tenants
SET name = $2, plan_id = $3, settings = $4, updated_at = now()
WHERE id = $1;

-- name: DisableTenant :exec
UPDATE tenants SET enabled = false, updated_at = now() WHERE id = $1;

-- name: EnableTenant :exec
UPDATE tenants SET enabled = true, updated_at = now() WHERE id = $1;
