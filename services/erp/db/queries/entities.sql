-- name: ListEntities :many
SELECT id, tenant_id, type, code, name, encrypted_tax_id, tax_id_hash,
       email, phone, address, metadata, active, created_at, updated_at
FROM erp_entities
WHERE tenant_id = $1
  AND type = $2
  AND deleted_at IS NULL
  AND (sqlc.arg(active_only)::BOOLEAN = false OR active = true)
  AND (sqlc.arg(search)::TEXT = '' OR name ILIKE '%' || sqlc.arg(search)::TEXT || '%'
       OR code ILIKE '%' || sqlc.arg(search)::TEXT || '%')
ORDER BY name
LIMIT $3 OFFSET $4;

-- name: CountEntities :one
SELECT COUNT(*)::INT AS count
FROM erp_entities
WHERE tenant_id = $1
  AND type = $2
  AND deleted_at IS NULL
  AND (sqlc.arg(active_only)::BOOLEAN = false OR active = true);

-- name: GetEntity :one
SELECT id, tenant_id, type, code, name, encrypted_tax_id, tax_id_hash,
       email, phone, address, metadata, active, created_at, updated_at
FROM erp_entities
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: CreateEntity :one
INSERT INTO erp_entities (tenant_id, type, code, name, encrypted_tax_id, tax_id_hash, email, phone, address, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, tenant_id, type, code, name, encrypted_tax_id, tax_id_hash, email, phone, address, metadata, active, created_at, updated_at;

-- name: UpdateEntity :one
UPDATE erp_entities
SET code = $3, name = $4, encrypted_tax_id = $5, tax_id_hash = $6,
    email = $7, phone = $8, address = $9, metadata = $10, active = $11, updated_at = now()
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
RETURNING id, tenant_id, type, code, name, encrypted_tax_id, tax_id_hash, email, phone, address, metadata, active, created_at, updated_at;

-- name: SoftDeleteEntity :exec
UPDATE erp_entities SET deleted_at = now(), active = false, updated_at = now()
WHERE id = $1 AND tenant_id = $2;

-- name: ListEntityContacts :many
SELECT id, tenant_id, entity_id, type, label, value, metadata, created_at
FROM erp_entity_contacts
WHERE entity_id = $1 AND tenant_id = $2
ORDER BY created_at;

-- name: CreateEntityContact :one
INSERT INTO erp_entity_contacts (tenant_id, entity_id, type, label, value, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tenant_id, entity_id, type, label, value, metadata, created_at;

-- name: DeleteEntityContact :exec
DELETE FROM erp_entity_contacts WHERE id = $1 AND tenant_id = $2;

-- name: ListEntityDocuments :many
SELECT id, tenant_id, entity_id, name, doc_type, file_key, uploaded_at
FROM erp_entity_documents
WHERE entity_id = $1 AND tenant_id = $2
ORDER BY uploaded_at DESC;

-- name: CreateEntityDocument :one
INSERT INTO erp_entity_documents (tenant_id, entity_id, name, doc_type, file_key)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, entity_id, name, doc_type, file_key, uploaded_at;

-- name: DeleteEntityDocument :exec
DELETE FROM erp_entity_documents WHERE id = $1 AND tenant_id = $2;

-- name: ListEntityNotes :many
SELECT id, tenant_id, entity_id, user_id, type, body, created_at
FROM erp_entity_notes
WHERE entity_id = $1 AND tenant_id = $2
ORDER BY created_at DESC
LIMIT $3;

-- name: CreateEntityNote :one
INSERT INTO erp_entity_notes (tenant_id, entity_id, user_id, type, body)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, entity_id, user_id, type, body, created_at;

-- name: ListEntityRelations :many
SELECT id, tenant_id, from_id, to_id, type, metadata, created_at
FROM erp_entity_relations
WHERE (from_id = $1 OR to_id = $1) AND tenant_id = $2
ORDER BY created_at;
