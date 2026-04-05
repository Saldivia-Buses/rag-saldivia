-- name: ListModules :many
SELECT id, name, category, description, icon, version, requires, tier_min, enabled
FROM modules
WHERE enabled = true
ORDER BY category, name;

-- name: GetEnabledModulesForTenant :many
SELECT m.id, m.name, m.category, m.icon, tm.config, tm.enabled_at
FROM tenant_modules tm
JOIN modules m ON m.id = tm.module_id
WHERE tm.tenant_id = $1 AND tm.enabled = true AND m.enabled = true
ORDER BY m.category, m.name;

-- name: EnableModuleForTenant :exec
INSERT INTO tenant_modules (tenant_id, module_id, enabled, config, enabled_by)
VALUES ($1, $2, true, $3, $4)
ON CONFLICT (tenant_id, module_id) DO UPDATE
SET enabled = true, config = $3, enabled_at = now(), enabled_by = $4;

-- name: DisableModuleForTenant :exec
UPDATE tenant_modules SET enabled = false WHERE tenant_id = $1 AND module_id = $2;

-- name: GetModuleConfigForTenant :one
SELECT config FROM tenant_modules
WHERE tenant_id = $1 AND module_id = $2 AND enabled = true;
