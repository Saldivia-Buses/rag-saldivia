-- ─── Tools (Phase 1 §Data migration — Pareto #4) ──────────────────────────
-- HERRAMIENTAS → erp_tools and HERRMOVS → erp_tool_movements. Despite the
-- "herramientas" naming, HERRAMIENTAS is the serialized inventory tag
-- ledger; HERRMOVS is the lending ledger for tools/items assigned to
-- employees.

-- name: ListTools :many
-- Serialized tool/item catalog. Filters by status and optional article code.
SELECT id, tenant_id, legacy_id, code, article_code, article_id,
       inventory_code, name, characteristic, group_code, tool_type,
       status_code, purchase_order_no, purchase_order_date,
       delivery_note_date, delivery_note_post, delivery_note_no,
       supplier_code, pending_oc, observation, manufacture_no,
       generated_at, created_at
FROM erp_tools
WHERE tenant_id = $1
  AND (sqlc.arg(status_filter)::INTEGER = -1 OR status_code = sqlc.arg(status_filter)::INTEGER)
  AND (sqlc.arg(article_filter)::TEXT = '' OR article_code = sqlc.arg(article_filter)::TEXT)
ORDER BY name, code
LIMIT $2 OFFSET $3;

-- name: GetTool :one
SELECT id, tenant_id, legacy_id, code, article_code, article_id,
       inventory_code, name, characteristic, group_code, tool_type,
       status_code, purchase_order_no, purchase_order_date,
       delivery_note_date, delivery_note_post, delivery_note_no,
       supplier_code, pending_oc, observation, manufacture_no,
       generated_at, created_at
FROM erp_tools
WHERE id = $1 AND tenant_id = $2;

-- name: ListToolMovements :many
-- Lending ledger for a single tool (by tool_id). Returns the movement
-- history most-recent first; tool_id nullable so orphan movements are
-- also visible via ListToolMovementsByCode below.
SELECT id, tenant_id, legacy_id, tool_id, tool_code, user_code,
       quantity, movement_date, concept_code, created_at
FROM erp_tool_movements
WHERE tenant_id = $1 AND tool_id = $2
ORDER BY movement_date DESC NULLS LAST, legacy_id DESC
LIMIT $3 OFFSET $4;
