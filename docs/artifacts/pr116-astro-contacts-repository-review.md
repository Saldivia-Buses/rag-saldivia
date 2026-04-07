# Gateway Review -- PR #116 Astro Contacts Repository (Phase 12)

**Fecha:** 2026-04-06
**Resultado:** CAMBIOS REQUERIDOS

## Bloqueantes

### 1. [queries.sql:2-3] GetContact/DeleteContact missing user_id filter -- IDOR

`GetContact` and `DeleteContact` filter only by `tenant_id + id`. Any authenticated user within the same tenant can read or delete another user's contacts by guessing/enumerating the UUID.

**Fix:** Add `AND user_id = $3` to both queries:

```sql
-- name: GetContact :one
SELECT * FROM contacts WHERE tenant_id = $1 AND id = $2 AND user_id = $3;

-- name: DeleteContact :exec
DELETE FROM contacts WHERE tenant_id = $1 AND id = $2 AND user_id = $3;
```

### 2. [queries.sql:5-6] GetContactByName missing user_id filter -- same IDOR

`GetContactByName` filters by `tenant_id + lower(name)` but not `user_id`. Cross-user data leak within the same tenant.

**Fix:** Add `AND user_id = $3`:

```sql
-- name: GetContactByName :one
SELECT * FROM contacts WHERE tenant_id = $1 AND user_id = $2 AND lower(name) = lower($3);
```

### 3. [queries.sql:19-23] UpdateContact missing user_id filter -- same IDOR

`UpdateContact` WHERE clause is `tenant_id = $1 AND id = $2` only. User A can overwrite User B's contact.

**Fix:** Add `AND user_id = $3` to WHERE (shift param numbers accordingly).

## Debe corregirse

### 4. [010_astro_contacts.up.sql:22] UNIQUE INDEX is tenant-scoped, should be tenant+user scoped

`idx_contacts_tenant_name` is `UNIQUE(tenant_id, lower(name))`. This means two different users in the same tenant cannot have a contact with the same name. The design intent (per the `user_id` column and ListContacts query) is that contacts are per-user.

**Fix:**
```sql
CREATE UNIQUE INDEX idx_contacts_tenant_user_name ON contacts(tenant_id, user_id, lower(name));
```

### 5. [queries.sql.go:219] SearchContacts param name is `Column3`

sqlc generated `Column3 pgtype.Text` because the SQL expression `'%' || $3 || '%'` is unnamed. This is cosmetic but makes the handler code ugly. Fix by using a named cast in the SQL:

```sql
-- name: SearchContacts :many
SELECT * FROM contacts WHERE tenant_id = $1 AND user_id = $2 AND name ILIKE '%' || @query::text || '%' ORDER BY name;
```

This will generate `Query string` as the field name instead of `Column3`.

### 6. [queries.sql] Missing LIMIT/OFFSET on ListContacts and SearchContacts

Both return unbounded result sets. A user with thousands of contacts will get them all in one query. Add pagination:

```sql
-- name: ListContacts :many
SELECT * FROM contacts WHERE tenant_id = $1 AND user_id = $2 ORDER BY name LIMIT $3 OFFSET $4;
```

## Sugerencias

- **Missing `CountContacts` query** -- the handler will likely need a total count for pagination. Add `SELECT count(*) FROM contacts WHERE tenant_id = $1 AND user_id = $2`.
- **Down migration is fine** -- `DROP TABLE IF EXISTS contacts` is correct and present.
- **db.go is clean** -- standard sqlc-generated DBTX interface + WithTx, no issues.
- **models.go** -- Contact struct is correct. Note that `models.go` also contains all other tenant models (Session, Message, etc.) because sqlc reads all migrations in `db/tenant/migrations/`. This is expected behavior with the shared schema directory.
- **sqlc.yaml** points to the shared tenant migrations directory, which is correct for the multi-tenant single-schema approach.

## Lo que esta bien

- **Tenant isolation at SQL level**: Every query includes `tenant_id = $1` as first filter -- correct pattern.
- **No raw SQL in Go**: All queries are sqlc-generated with parameterized placeholders. No injection risk.
- **Migration structure**: UP/DOWN pair present, UUID PK with `gen_random_uuid()`, proper indexes on `tenant_id` and `(tenant_id, user_id)`.
- **Nullable fields**: `relationship`, `notes`, `birth_time` are correctly nullable (`pgtype.Text`, `pgtype.Time`).
- **Sensible defaults**: `nation='Argentina'`, `utc_offset=-3`, `alt=25.0`, `kind='persona'` match the astro domain.
- **Handler is still stubs** -- no risk of shipping broken wiring yet. The IDOR fixes must be applied before the handler implementation PR.
