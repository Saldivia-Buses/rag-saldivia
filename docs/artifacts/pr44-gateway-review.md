# Gateway Review -- PR #44: Database Bootstrapping Layer

**Fecha:** 2026-04-02
**Tipo:** review
**Branch:** `feat/migrations-seed-cli`
**Intensity:** thorough

## Resultado

**CAMBIOS REQUERIDOS** (1 bloqueante, 3 debe corregirse, 5 sugerencias)

---

## Hallazgos

### Bloqueantes

1. **[deploy/scripts/seed.sh:52] Heredoc expansion corrupts bcrypt hashes**

   The tenant DB seed uses an unquoted heredoc delimiter `<<SQL` (line 52), which
   triggers bash variable expansion. The bcrypt hashes contain `$2b`, `$12`,
   `$EC4H`, etc. -- bash interprets each `$...` segment as a variable reference
   (all empty), producing a mangled string like
   `EC4HSJuWfO8E52aGBICWbuy/g3HSEMk2L2ml6N7DNfWUriOm0XTV.` instead of the full
   hash. **Login will fail for all seeded users.**

   Contrast with line 17 which correctly uses `<<'SQL'` (quoted, no expansion)
   for the platform seed.

   **Fix:** Change line 52 from `<<SQL` to `<<'SQL'`, and pass the hashes via
   psql variables instead:

   ```bash
   psql "$TENANT_DB_URL" -v ON_ERROR_STOP=1 --quiet \
       -v admin_hash="$ADMIN_HASH" \
       -v user_hash="$USER_HASH" \
       <<'SQL'
   INSERT INTO users (id, email, name, password_hash)
   VALUES ('u-admin', 'admin@sda.local', 'Enzo Saldivia', :'admin_hash')
   ON CONFLICT (email) DO NOTHING;

   INSERT INTO users (id, email, name, password_hash)
   VALUES ('u-user', 'user@sda.local', 'Usuario Test', :'user_hash')
   ON CONFLICT (email) DO NOTHING;

   INSERT INTO user_roles (user_id, role_id)
   VALUES ('u-admin', 'role-admin')
   ON CONFLICT DO NOTHING;

   INSERT INTO user_roles (user_id, role_id)
   VALUES ('u-user', 'role-user')
   ON CONFLICT DO NOTHING;
   SQL
   ```

   The `:'varname'` syntax passes psql variables as properly quoted literals.

---

### Debe corregirse

2. **[deploy/scripts/migrate.sh:36] `|| true` silently swallows migration failures**

   The `run_sql` function pipes psql through `grep -v "^$" || true`. Combined
   with `set -o pipefail`, the `|| true` catches any pipeline failure -- meaning
   if psql exits non-zero (via `ON_ERROR_STOP=1`), the error is silently
   swallowed and the script continues to the next migration as if nothing
   happened. A broken migration will not stop the chain.

   **Fix:** Separate the error filtering from the error propagation:

   ```bash
   run_sql() {
       local db_url="$1"
       local file="$2"
       local name
       name=$(basename "$(dirname "$(dirname "$(dirname "$file")")")")
       log "applying $name/$(basename "$file") -> $(echo "$db_url" | sed 's|.*@||; s|?.*||')"
       psql "$db_url" -f "$file" -v ON_ERROR_STOP=1 --quiet 2>&1 | grep -v "^$" || {
           local rc=${PIPESTATUS[0]}
           if [ "$rc" -ne 0 ]; then
               log "FAILED (exit $rc): $name/$(basename "$file")"
               exit "$rc"
           fi
       }
   }
   ```

   Alternatively, simpler approach -- drop the grep entirely since `--quiet`
   already suppresses most noise:

   ```bash
   psql "$db_url" -f "$file" -v ON_ERROR_STOP=1 --quiet
   ```

3. **[deploy/docker-compose.dev.yml:222-228] `notification` service loses `db-init` dependency**

   The `notification` service uses `<<: *service-defaults` (which includes
   `depends_on: db-init`) but then explicitly overrides `depends_on` at
   lines 222-228 to add `mailpit`. In Docker Compose YAML merge, the explicit
   `depends_on` completely replaces the anchor's `depends_on` -- so `db-init`
   is dropped. If `notification` starts before migrations complete, it will fail
   on missing tables.

   **Fix:** Include `db-init` in the explicit `depends_on`:

   ```yaml
   depends_on:
     db-init:
       condition: service_completed_successfully
     postgres:
       condition: service_healthy
     nats:
       condition: service_healthy
     mailpit:
       condition: service_started
   ```

   Same pattern applies to `platform` (lines 248-249) -- it depends on
   `postgres` directly but not on `db-init`. Since `platform` queries the
   platform DB that `db-init` populates, it should also depend on `db-init`:

   ```yaml
   platform:
     depends_on:
       db-init:
         condition: service_completed_successfully
       postgres:
         condition: service_healthy
   ```

4. **[deploy/scripts/seed.sh:25] Seed stores tenant `postgres_url` with password in platform DB**

   The seed inserts the dev tenant's full connection string including password
   (`postgres://sda:sda_dev@postgres:5432/...`) into the `tenants.postgres_url`
   column. In the dev environment this is expected, but this is the pattern that
   production code will follow. The bible says "Secrets en Docker secrets o
   Vault, nunca en env vars planos" -- storing credentials in a DB column is the
   same class of issue.

   **Fix for now:** Add a comment in seed.sh marking this as dev-only and noting
   that production tenant onboarding must resolve connection strings from a
   secrets manager, not store credentials inline. Consider splitting to
   `postgres_host` + secret reference in the schema for production.

---

### Sugerencias

5. **[deploy/scripts/seed.sh:49-50] Both users share the same password and hash**

   `ADMIN_HASH` and `USER_HASH` are identical -- same password "admin123" for
   both roles. For dev testing this is fine, but having distinct passwords per
   role makes it easier to verify RBAC in manual testing (you know which role you
   logged in as by the password you typed). Low priority -- just a dev ergonomics
   note.

6. **[services/auth/db/migrations/001_init.up.sql:85-138] Seed data in migration files**

   The auth migration contains both DDL (CREATE TABLE) and DML (INSERT INTO
   roles, permissions, role_permissions). Mixing schema definition with seed data
   in the same migration makes it harder to write `001_init.down.sql` and
   complicates future schema-only tooling (e.g., diffing schema between
   environments). Consider extracting the INSERT statements to a separate
   `002_seed_roles.up.sql`. Same pattern in `platform/001_init.up.sql`
   (lines 111-157).

   That said -- for system-level seed data (roles, plans, modules) that is
   required for the app to function, embedding it in the migration is defensible.
   Just be consistent about the distinction: "structural seed data" (roles, plans)
   goes in migrations, "test data" (users, tenants) goes in seed.sh.

7. **[deploy/docker-compose.dev.yml:113] `db-init` container uses `postgres:16-alpine` image (230MB)**

   The `db-init` container only needs `psql` and `bash`. Using the full
   `postgres:16-alpine` image works but is heavier than necessary. This is
   a minor optimization -- could use a smaller image with just the psql client,
   but the convenience of matching the server image version outweighs the size
   concern for a dev stack. No action needed.

8. **[deploy/scripts/migrate.sh, seed.sh] Scripts should be executable**

   The Makefile targets `migrate` and `seed` invoke the scripts directly (not
   via `bash script.sh`). If the scripts lack `+x` permission, the host-side
   `make migrate` and `make seed` will fail with "Permission denied". The Docker
   path uses `bash -c` so it works regardless. Verify that `chmod +x` was
   included in the commit.

9. **[Makefile:107] `migrate-seed` is a nice compound target**

   The `migrate-seed: migrate seed` pattern is clean. Consider also adding a
   `db-reset` target that drops and recreates the databases before running
   migrate-seed, for when devs need a clean slate:

   ```makefile
   db-reset: ## Drop + recreate dev databases, then migrate + seed
   	psql "postgres://sda:sda_dev@localhost:5432/postgres" -c "DROP DATABASE IF EXISTS sda_platform; DROP DATABASE IF EXISTS sda_tenant_dev;"
   	psql "postgres://sda:sda_dev@localhost:5432/postgres" -f $(DEPLOY_DIR)/postgres-init.sql
   	$(MAKE) migrate-seed
   ```

---

### Lo que esta bien

- **Migration idempotency is solid.** Every CREATE TABLE uses `IF NOT EXISTS`,
  every INSERT uses `ON CONFLICT DO NOTHING`. Re-running is safe. This is the
  correct pattern for a `db-init` container that runs on every `docker compose up`.

- **Correct FK dependency order in migrate.sh.** The `TENANT_MIGRATION_ORDER`
  array explicitly encodes `auth -> chat -> notification -> ingest`, which
  matches the FK graph (all tenant services reference `users(id)` from auth).
  Platform runs separately against its own DB. Well thought out.

- **Credential stripping in logs.** The `sed 's|.*@||; s|?.*||'` in `run_sql`
  strips user:password from the DB URL before logging. Produces
  `postgres:5432/sda_tenant_dev` -- clean output, no credential leak.

- **Container path resolution is robust.** The `if [ -d "/services" ]` check in
  migrate.sh correctly distinguishes between running inside Docker (where services
  are mounted at `/services`) vs running on the host (where it navigates relative
  to the script). Clean dual-mode support.

- **`set -euo pipefail` in both scripts.** Proper shell hardening. Combined with
  `ON_ERROR_STOP=1` on psql, this is the right approach (modulo the `|| true`
  issue noted above).

- **Platform seed uses quoted heredoc `<<'SQL'`.** Correct pattern that prevents
  shell expansion. Just needs to be applied consistently to the tenant seed too.

- **Docker Compose structure is clean.** The `x-service-defaults` anchor with
  `service_completed_successfully` on `db-init` ensures services wait for
  migrations. The `restart: "no"` on `db-init` is correct for a run-once init
  container. The `profiles: [full]` split between infra-only and full mode is
  well designed.

- **Schema design is production-quality.** Proper use of TIMESTAMPTZ, JSONB
  defaults, CHECK constraints, cascade deletes, partial indexes
  (`WHERE status IN ('pending', 'processing')`), and composite primary keys.
  The auth schema covers brute force lockout, MFA, and immutable audit log.

---

## Security Notes

- **bcrypt cost 12** is appropriate for dev seeds. Production should use cost 12-14.
- **`JWT_SECRET: dev-secret-at-least-32-characters-long!!`** in env-common is
  fine for dev. The `at-least-32-characters-long` self-documenting name is a nice
  touch.
- **`PGPASSWORD` in db-init environment** (line 123) is needed for psql but is
  only in the dev compose. Acceptable.
- **No `.env` file with real secrets is committed.** The `.env.example` has dev
  defaults only. Good.
