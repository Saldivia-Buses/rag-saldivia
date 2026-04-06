-- name: GetContact :one
SELECT * FROM contacts WHERE tenant_id = $1 AND id = $2;

-- name: GetContactByName :one
SELECT * FROM contacts WHERE tenant_id = $1 AND lower(name) = lower($2);

-- name: ListContacts :many
SELECT * FROM contacts WHERE tenant_id = $1 AND user_id = $2 ORDER BY name;

-- name: SearchContacts :many
SELECT * FROM contacts WHERE tenant_id = $1 AND user_id = $2 AND name ILIKE '%' || $3 || '%' ORDER BY name;

-- name: CreateContact :one
INSERT INTO contacts (tenant_id, user_id, name, birth_date, birth_time, birth_time_known, city, nation, lat, lon, alt, utc_offset, relationship, notes, kind)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING *;

-- name: UpdateContact :one
UPDATE contacts SET
    name = $3, birth_date = $4, birth_time = $5, birth_time_known = $6,
    city = $7, nation = $8, lat = $9, lon = $10, alt = $11, utc_offset = $12,
    relationship = $13, notes = $14, kind = $15, updated_at = now()
WHERE tenant_id = $1 AND id = $2
RETURNING *;

-- name: DeleteContact :exec
DELETE FROM contacts WHERE tenant_id = $1 AND id = $2;
