# Gateway Review -- Plan 07 Consolidation (Phases 1, 2a, 4)

**Fecha:** 2026-04-05
**Resultado:** CAMBIOS REQUERIDOS

**PR:** #82
**Branch:** `feat/plan07-consolidation`
**Scope:** Centralized migrations, `pkg/config`, NATS standardization

---

## Bloqueantes

### B1. NATS Notify still does not validate event type -- subject injection possible
**File:** `/home/enzo/rag-saldivia/pkg/nats/publisher.go:82-90`

The `Notify` method validates `tenantSlug` via `IsValidSubjectToken` but does NOT validate `parsed.Type` before interpolating it into the NATS subject:

```go
subject := "tenant." + tenantSlug + ".notify." + parsed.Type
```

If `parsed.Type` contains NATS wildcards (`*`, `>`), spaces, or control characters, the resulting subject is malformed. While NATS server rejects publishing to wildcard subjects, relying solely on server-side rejection is not defense-in-depth.

Dots are intentional in the type (e.g., `chat.new_message`), so `IsValidSubjectToken` (which rejects dots) cannot be reused. A new validation function is needed.

**Fix:** Add `IsValidEventType(s string) bool` with regex `^[a-zA-Z0-9_][a-zA-Z0-9_.-]*$` (alphanumeric, underscore, hyphen, dot -- no wildcards, spaces, or control chars). Call it before subject construction:

```go
if !IsValidEventType(parsed.Type) {
    return fmt.Errorf("invalid event type for NATS subject: %q", parsed.Type)
}
```

This was flagged in PR #37, PR #52, and every NATS review since. This PR explicitly adds `IsValidSubjectToken` as the canonical validation, but omits event-type validation. The gap must be closed in this PR since it claims to standardize NATS publishing.

---

### B2. Old migration files in `services/*/db/migrations/` not removed -- creates confusion about source of truth

The PR adds centralized migrations under `db/tenant/migrations/` and `db/platform/migrations/` but leaves all the old migration files in place:

- `services/auth/db/migrations/` -- 001, 002
- `services/chat/db/migrations/` -- 000_deps, 001, 002
- `services/notification/db/migrations/` -- 000_deps, 001
- `services/ingest/db/migrations/` -- 000_deps, 001, 002
- `services/platform/db/migrations/` -- 001, 002, 003, 004
- `services/feedback/db/migrations/` -- 001

Two migration directories with the same content creates a maintenance hazard: a developer edits one and not the other, and now runtime differs from what sqlc generates code against.

**Fix:** Choose one of:

**(A) Symlink (recommended):** Per-service `db/migrations/` dirs become symlinks to `../../db/{tenant|platform}/migrations/`. sqlc reads the symlink, migrate.sh reads the real dir. Single source of truth. The `000_deps.up.sql` stubs would move into the centralized dir (harmless since `CREATE TABLE IF NOT EXISTS` is idempotent, or they could be dropped since the full `users` table is in `001_auth_init`).

**(B) sqlc overrides:** Update each `sqlc.yaml` to point `schema:` at the centralized `db/` dir. Requires relative path like `../../../../db/tenant/migrations/` which is fragile but works.

**(C) Explicit documentation + CI check:** Keep duplicates but add a CI step that diffs them and fails if they diverge. Least clean but least disruptive.

Whichever approach: the PR description should document the chosen strategy.

---

## Debe corregirse

### C1. `migrate.sh` -- no transaction wrapping around migration application
**File:** `/home/enzo/rag-saldivia/deploy/scripts/migrate.sh:51-54`

```bash
psql "$db_url" -f "$file" -v ON_ERROR_STOP=1 --quiet
psql "$db_url" --quiet -c "INSERT INTO schema_migrations ..."
```

If psql succeeds on the migration file but the script crashes before the INSERT into `schema_migrations`, the next run will try to apply the migration again. For most files this is safe (IF NOT EXISTS), but `002_auth_audit_actions.up.sql` uses:

```sql
ALTER TABLE audit_log ADD CONSTRAINT audit_log_action_format ...
```

This has no `IF NOT EXISTS` guard. A retry would fail with "constraint already exists", blocking all subsequent migrations.

**Fix:** Wrap the migration + tracking in a single psql transaction:

```bash
apply_migration() {
    local db_url="$1"
    local file="$2"
    local filename
    filename=$(basename "$file")

    psql "$db_url" -v ON_ERROR_STOP=1 --quiet <<SQL
BEGIN;
SELECT pg_advisory_xact_lock(hashtext('$filename'));
DO \$\$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM schema_migrations WHERE filename = '$filename') THEN
    \i $file
    INSERT INTO schema_migrations (filename) VALUES ('$filename');
  END IF;
END \$\$;
COMMIT;
SQL
}
```

If that's too complex for a bash script, at minimum make `002_auth_audit_actions.up.sql` idempotent by checking `pg_constraint` before adding the constraint, or use `DO $$ BEGIN ... EXCEPTION WHEN duplicate_object THEN NULL; END $$;`.

### C2. `migrate.sh` -- check-then-act race condition
**File:** `/home/enzo/rag-saldivia/deploy/scripts/migrate.sh:42-54`

The check (SELECT) and act (migration + INSERT) are not atomic. Two concurrent `migrate.sh` processes (e.g., docker-compose scaling, or parallel dev workflows) would both see "not applied" and both try to apply.

For DDL with `IF NOT EXISTS` this is safe. For `INSERT ... ON CONFLICT DO NOTHING` in seed data this is safe. But `ALTER TABLE ADD CONSTRAINT` (migration 002) is not idempotent -- the second process would fail.

**Fix:** Use `pg_advisory_lock` or wrap in a transaction with `SELECT FOR UPDATE` (as in C1). Alternatively, document that `migrate.sh` must not be run concurrently (simplest fix).

### C3. `docker-compose.dev.yml` -- dead volume mount `../services:/services:ro` in db-init
**File:** `/home/enzo/rag-saldivia/deploy/docker-compose.dev.yml:229`

The `db-init` service mounts:
```yaml
volumes:
  - ./scripts:/scripts:ro
  - ../db:/db:ro
  - ../services:/services:ro    # <-- dead
```

`migrate.sh` now reads from `/db` (the centralized dir). `seed.sh` does not reference `/services` either. The `../services:/services:ro` mount is leftover from the old per-service migration approach.

**Fix:** Remove the `../services:/services:ro` line.

### C4. `migrate.sh` -- SQL injection via filename interpolation
**File:** `/home/enzo/rag-saldivia/deploy/scripts/migrate.sh:44,54`

```bash
applied=$(psql "$db_url" -t -c "SELECT 1 FROM schema_migrations WHERE filename = '$filename'" ...)
psql "$db_url" --quiet -c "INSERT INTO schema_migrations (filename) VALUES ('$filename') ..."
```

`$filename` comes from `basename "$file"` (filesystem path), so it's safe in practice. But if a migration file is ever named with a single quote (unlikely but possible), this becomes SQL injection.

**Fix:** Use psql variables instead of shell interpolation:

```bash
applied=$(psql "$db_url" -t -v fn="$filename" \
  -c "SELECT 1 FROM schema_migrations WHERE filename = :'fn'" 2>/dev/null | tr -d ' ')
```

And for the insert:
```bash
psql "$db_url" --quiet -v fn="$filename" \
  -c "INSERT INTO schema_migrations (filename) VALUES (:'fn') ON CONFLICT DO NOTHING"
```

### C5. Missing test: NATS Notify with wildcard in event type
**File:** `/home/enzo/rag-saldivia/pkg/nats/publisher_test.go`

The test suite tests invalid slugs thoroughly but has no test for invalid event types (wildcards, empty, spaces). Once B1 is fixed with `IsValidEventType`, add:

```go
func TestPublisher_Notify_InvalidType_Returns_Error(t *testing.T) {
    nc := startTestNATS(t)
    pub := New(nc)
    tests := []struct{ name, typ string }{
        {"wildcard_star", "chat.*"},
        {"wildcard_gt", "chat.>"},
        {"spaces", "chat new_message"},
        {"empty", ""},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := pub.Notify("saldivia", Event{Type: tt.typ, UserID: "u-1", Title: "t", Body: "b"})
            if err == nil {
                t.Errorf("expected error for type %q", tt.typ)
            }
        })
    }
}
```

---

## Sugerencias

### S1. `pkg/config/env.go` -- consider `t.Setenv` in tests

The tests use `os.Setenv` + `defer os.Unsetenv`. Since Go 1.17+, `t.Setenv` does this automatically and is test-parallel-safe:

```go
func TestEnv_WithValue(t *testing.T) {
    t.Setenv("TEST_ENV_KEY", "hello")
    got := config.Env("TEST_ENV_KEY", "default")
    if got != "hello" {
        t.Fatalf("expected hello, got %s", got)
    }
}
```

### S2. `pkg/config/env.go` -- MustEnv edge case with whitespace-only values

`MustEnv` only checks `v == ""`. An env var set to `"  "` (whitespace only) would pass silently. This could be intentional (some legitimate values are whitespace-sensitive), but for infrastructure config like DB URLs it's likely a bug. Consider `strings.TrimSpace(v) == ""` or leave as-is with a doc comment explaining the behavior.

### S3. Centralized migration numbering -- document the renumbering strategy

The tenant migrations are renumbered from per-service numbering (auth 001/002, chat 001/002, etc.) to global sequential (001-008). The mapping should be documented somewhere (even a comment in the migration dir or in the PR description) so future developers understand the provenance:

```
001 = auth/001_init
002 = auth/002_audit_actions
003 = chat/001_init
004 = chat/002_add_thinking
005 = notification/001_init
006 = ingest/001_init
007 = ingest/002_intelligence_schema (plan06)
008 = feedback/001_init
```

### S4. `migrate.sh` -- add `--single-transaction` flag to psql

For DDL-heavy migrations, `psql --single-transaction -f "$file"` would wrap the entire file in one transaction. If any statement fails, the whole migration rolls back cleanly. This is safer than `ON_ERROR_STOP` which stops on error but leaves partial state.

### S5. NATS `Connect()` -- consider adding a name option

Adding `nats.Name("auth-service")` (or accepting it as a parameter) to the `Connect()` factory would make NATS monitoring more useful. Each connection would be identifiable in the NATS dashboard (`nats server list`).

```go
func Connect(url, name string) (*nats.Conn, error) {
    return nats.Connect(url,
        nats.Name(name),
        // ... existing options
    )
}
```

### S6. `pkg/config` adoption plan

None of the 10 services with `func env()` have been migrated to use `pkg/config.Env()`. This is fine if planned for a future phase, but should be called out in the PR description to avoid the impression that the refactor is complete.

---

## Lo que esta bien

1. **Migration consolidation strategy is sound.** Having one ordered sequence per database (tenant 001-008, platform 001-004) eliminates the FK dependency nightmare where chat's `000_deps.up.sql` had to create a stub `users` table.

2. **`schema_migrations` tracking is the right approach.** Simple, filesystem-name-based tracking is better than versioned integers (like golang-migrate) for a project where migrations are renamed/renumbered during consolidation.

3. **`IsValidSubjectToken` is now exported and well-tested.** 14 test cases covering empty, dots, wildcards, whitespace, control characters. The allowlist approach (`^[a-zA-Z0-9_-]+$`) is correct.

4. **NATS `Connect()` factory centralizes reconnect behavior.** `RetryOnFailedConnect(true)`, `MaxReconnects(-1)`, and disconnect/reconnect logging via slog are all correct defaults. Services currently roll their own connect options with inconsistent settings (some have reconnect logging, some don't).

5. **`Broadcast` now validates both slug AND channel.** This was flagged as a bug in PR #37 and is now fixed.

6. **`pkg/config` is clean and minimal.** `Env()` and `MustEnv()` -- no over-abstraction, no magic, does exactly what it says.

7. **All migration UP files have matching DOWN files.** Every `.up.sql` has a corresponding `.down.sql` with correct reverse ordering (drop tables in reverse dependency order).

8. **Seed data uses `ON CONFLICT DO NOTHING` everywhere.** Safe to re-run.

9. **docker-compose.dev.yml correctly mounts `../db:/db:ro`.** The `db-init` container can see the centralized migration directory at `/db`.

---

## Summary

Two bloqueantes: (B1) NATS event type injection via Notify is the same gap flagged since PR #37 and must be closed in this PR since it explicitly claims to standardize NATS publishing; (B2) duplicate migration files in `services/` and `db/` with no strategy for keeping them in sync or documenting why both exist.

Five must-fix items: non-atomic migration application, race condition, dead volume mount, SQL interpolation in migrate.sh, and missing tests for event type validation.

Six suggestions for polish.

Overall this is good foundation work. The migration consolidation simplifies the system, `pkg/config` is clean, and the NATS improvements fix real bugs. The blockers are all fixable within the scope of this PR.
