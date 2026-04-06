-- name: GetContact :one
SELECT * FROM contacts WHERE tenant_id = $1 AND user_id = $2 AND id = $3;

-- name: GetContactByName :one
SELECT * FROM contacts WHERE tenant_id = $1 AND user_id = $2 AND lower(name) = lower($3);

-- name: ListContacts :many
SELECT * FROM contacts WHERE tenant_id = $1 AND user_id = $2 ORDER BY name LIMIT $3 OFFSET $4;

-- name: SearchContacts :many
SELECT * FROM contacts WHERE tenant_id = $1 AND user_id = $2 AND name ILIKE '%' || @query::text || '%' ORDER BY name LIMIT $3 OFFSET $4;

-- name: CreateContact :one
INSERT INTO contacts (tenant_id, user_id, name, birth_date, birth_time, birth_time_known, city, nation, lat, lon, alt, utc_offset, relationship, notes, kind)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING *;

-- name: UpdateContact :one
UPDATE contacts SET
    name = $4, birth_date = $5, birth_time = $6, birth_time_known = $7,
    city = $8, nation = $9, lat = $10, lon = $11, alt = $12, utc_offset = $13,
    relationship = $14, notes = $15, kind = $16, updated_at = now()
WHERE tenant_id = $1 AND user_id = $2 AND id = $3
RETURNING *;

-- name: DeleteContact :exec
DELETE FROM contacts WHERE tenant_id = $1 AND user_id = $2 AND id = $3;
