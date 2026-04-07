# Gateway Review -- PR #87 Plan 08 Phase 1: Critical Backend Fixes

**Fecha:** 2026-04-05
**Branch:** `feat/plan08-phase-1` -> `2.0.1`
**Commits:** 5 (8561c6c, f182e3b, ab4ea38, ca5c53f, 287ff0a)
**Resultado:** CAMBIOS REQUERIDOS

---

## Bloqueantes

### B1. NATS container has no environment variables for per-service passwords
**File:** `deploy/docker-compose.prod.yml:124-141`

`deploy/nats/nats-server.conf` uses NATS-native `$VAR` syntax for password
interpolation (`$AUTH_NATS_PASS`, `$CHAT_NATS_PASS`, etc.). NATS server
resolves these from its own process environment. However, the NATS container
in docker-compose.prod.yml has **no `environment:` block** -- none of these
variables are passed to the container.

Result: NATS server resolves all passwords to empty strings. Every service
user has an empty password. Any client can connect as any user with an empty
password, which defeats the entire purpose of per-service auth.

**Fix:** Add environment variables to the NATS container:

```yaml
nats:
  environment:
    AUTH_NATS_PASS: ${AUTH_NATS_PASS}
    CHAT_NATS_PASS: ${CHAT_NATS_PASS}
    WS_NATS_PASS: ${WS_NATS_PASS}
    NOTIF_NATS_PASS: ${NOTIF_NATS_PASS}
    INGEST_NATS_PASS: ${INGEST_NATS_PASS}
    FEEDBACK_NATS_PASS: ${FEEDBACK_NATS_PASS}
    AGENT_NATS_PASS: ${AGENT_NATS_PASS}
    TRACES_NATS_PASS: ${TRACES_NATS_PASS}
    PLATFORM_NATS_PASS: ${PLATFORM_NATS_PASS}
    EXTRACTOR_NATS_PASS: ${EXTRACTOR_NATS_PASS}
```

Alternative (more secure): use NATS `include` directive with a mounted
secrets file so passwords don't appear in `docker inspect` output.

### B2. Platform service missing NATS_URL in docker-compose.prod.yml
**File:** `deploy/docker-compose.prod.yml:373-401`

The platform service block has no `NATS_URL` environment variable.
`services/platform/cmd/main.go:63` defaults to `nats://localhost:4222`,
which won't resolve inside Docker (NATS runs as container `nats` on the
`backend` network). Lifecycle events (tenant created, module changes)
will silently fail to publish.

The plan (C5) defines a `platform` NATS user with permissions, and the
`nats-server.conf` includes it, but the docker-compose wiring was not
completed.

**Fix:** Add to platform service environment:
```yaml
environment:
  PLATFORM_PORT: "8006"
  PLATFORM_TENANT_SLUG: platform
  NATS_URL: nats://platform:${PLATFORM_NATS_PASS}@nats:4222
```

### B3. Platform lifecycle events will be denied by NATS permissions
**Files:**
- `deploy/nats/nats-server.conf:84-91` (platform publish: `tenant.*.notify.*`)
- `services/platform/internal/service/platform.go:55-65` (event type construction)

`publishLifecycleEvent("tenant.created", ...)` calls `Notify()` with type
`"platform_tenant.created"`. The dot in the type is significant: `Notify()`
builds subject `tenant.{slug}.notify.platform_tenant.created` -- that is
**5 segments**. The NATS permission for platform is `tenant.*.notify.*`
(4 segments). The `*` wildcard matches exactly one token (no dots).

NATS will deny this publish with a permissions violation error.

This is a pre-existing issue in the publisher code, but this PR's NATS
auth config makes it actually enforced. Without fix, no platform lifecycle
events will ever be published in production.

**Fix (surgical, service-side):** In `publishLifecycleEvent`, sanitize the
event type to avoid dots:
```go
func (p *Platform) publishLifecycleEvent(tenantSlug, eventType string, data any) {
    if p.publisher == nil || tenantSlug == "" {
        return
    }
    safeType := "platform_" + strings.ReplaceAll(eventType, ".", "_")
    if err := p.publisher.Notify(tenantSlug, map[string]any{
        "type": safeType,
        "data": data,
    }); err != nil {
        slog.Warn("publish lifecycle event failed", "event", eventType, "error", err)
    }
}
```

This produces subject `tenant.{slug}.notify.platform_tenant_created` which
matches `tenant.*.notify.*`.

---

## Debe corregirse

### D1. Dev compose still has `rag` service with non-existent Dockerfile
**File:** `deploy/docker-compose.dev.yml:297-314`

C2 cleaned up Traefik configs and `docker-compose.prod.yml`, but the dev
compose still defines a `rag` service pointing to `services/rag/Dockerfile`
which does not exist (the `services/rag/` directory is empty/deleted).
Running `docker compose --profile full up --build` will fail.

**Fix:** Remove the `rag` service block (lines 297-314). Optionally add
agent, search, traces, feedback service entries for Docker-based dev.

### D2. Stale `nats_token` secret declaration
**File:** `deploy/docker-compose.prod.yml:25-26`

The `secrets:` block declares `nats_token` but no service references it.
Per-service NATS auth replaced the shared token. Dead configuration.

**Fix:** Remove the `nats_token` secret declaration (lines 25-26).

### D3. Secrets README documents stale NATS pattern
**File:** `deploy/secrets/README.md:16`

Still says:
```
| `nats-token` | nats, all Go services | Shared NATS auth token |
```

This is the old pattern. With per-service auth, operators need to set 10
environment variables (or provide them via `.env` file). No documentation
exists for the new credential model.

**Fix:** Replace the `nats-token` row with documentation for all 10
per-service NATS credentials, how to generate them, and how to provide
them to Docker compose (`.env` file or Docker secrets).

### D4. Platform migration 005 missing action format constraint
**File:** `db/platform/migrations/005_audit_log.up.sql`

The tenant `audit_log` has a CHECK constraint from migration 002:
```sql
ALTER TABLE audit_log ADD CONSTRAINT audit_log_action_format
    CHECK (action ~ '^[a-z]+\.[a-z_.]+$');
```

The platform `audit_log` (migration 005) has no such constraint. Since
both use the same `audit.Writer`, the same validation rules should apply.
Without the constraint, invalid action strings can be inserted into the
platform audit log without detection.

**Fix:** Add to migration 005:
```sql
ALTER TABLE audit_log ADD CONSTRAINT audit_log_action_format
    CHECK (action ~ '^[a-z]+\.[a-z_.]+$');
CREATE INDEX IF NOT EXISTS idx_platform_audit_log_action_created
    ON audit_log(action, created_at DESC);
```

### D5. `ensureSSLMode` fails silently on unparseable URLs
**File:** `pkg/tenant/resolver.go:268-271`

When `url.Parse` fails, the function returns the original URL without
logging. This means a malformed URL silently bypasses SSL enforcement
and fails later with a less clear error from pgxpool.

**Fix:**
```go
func ensureSSLMode(pgURL string) string {
    u, err := url.Parse(pgURL)
    if err != nil {
        slog.Warn("failed to parse PG URL for SSL enforcement", "error", err)
        return pgURL
    }
    // ... rest unchanged
}
```

This requires adding `"log/slog"` to the imports in `resolver.go`.

---

## Sugerencias

### S1. Consider Docker secrets for NATS passwords
The current approach uses Docker Compose `${VAR}` interpolation from
shell environment variables. These end up in `docker inspect` output
and the container's `/proc/1/environ`. Docker secrets (mounted as files)
would be more secure and consistent with how JWT keys and DB URLs are
handled. NATS supports `$include` for config fragments from files.

### S2. `ensureSSLMode` only protects resolver-based connections
**File:** `pkg/tenant/resolver.go:268-280`

Called in `createPoolLocked()` for tenant DB connections. But services
that connect directly via `pgxpool.New(ctx, dbURL)` in `main.go`
(platform, traces, feedback, search, ingest) bypass this entirely.
The real protection is the `sslmode=require` in the secrets files
(documented in README.md). Consider adding a comment in `main.go`
files noting that SSL is enforced at the secret level, not at code
level for direct connections.

### S3. Search service missing `backend` network
**File:** `deploy/docker-compose.prod.yml:296-305`

Search has networks `frontend` and `data` but not `backend`. While
search doesn't currently use NATS, if it ever needs to (e.g., Plan 08
Phase 4 could add broadcast for search results), it won't reach the
NATS container. Consider adding `backend` for consistency with other
services.

### S4. Add `ReadHeaderTimeout` to service configs
The plan (H-NEW in Fase 2) calls for `ReadHeaderTimeout: 10 * time.Second`
on all services (slowloris protection). No service currently has this.
Since this PR touches server configs in several services, consider adding
it here. This is technically Phase 2 scope but is a one-liner per service.

### S5. Explicit Traefik routing priority for feedback/platform overlap
**Files:** `deploy/traefik/dynamic/dev.yml:62,76` and `prod.yml:53,67`

Both `feedback` and `platform` routers match `/v1/platform/feedback/*`.
Traefik resolves by rule string length (longer wins). Currently works
because the feedback rule has `||` making it longer. But this implicit
priority is fragile. Adding explicit `priority:` fields would make the
intent clear:
```yaml
feedback:
  priority: 20
platform:
  priority: 10
```

### S6. Audit entries missing IP/UserAgent context
**Files:**
- `services/ingest/internal/service/ingest.go:175-178, 233-235`
- `services/platform/internal/service/platform.go:179-183, etc.`

New audit entries never populate `IP` and `UserAgent` fields. Auth
service populates these from the login request. The service layer
doesn't have access to the HTTP request. This reduces forensic value
but is not a security issue. Consider passing IP/UserAgent from handlers
to the service layer for consistency with auth's audit entries.

### S7. Platform `publishLifecycleEvent` only called on CreateTenant
**File:** `services/platform/internal/service/platform.go:178`

The function is only called once (CreateTenant). UpdateTenant, Disable/
EnableTenant, Enable/DisableModule, ToggleFeatureFlag, and SetConfig
do not emit lifecycle events. Other services that need to react to
module/config changes (e.g., auth's EnabledModules cache in Phase 4)
won't be notified. This is out of scope for Phase 1 but worth tracking.

---

## Lo que esta bien

### C1 -- Audit logging (ingest + platform)
- Clean integration following established `audit.NewWriter(pool)` pattern
- Ingest audits `ingest.upload` (with file/collection/size details) and
  `ingest.delete_job` -- exactly what the plan specified
- Platform audits all 8 mutation operations: tenant CRUD (4), module
  enable/disable (2), flag toggle (1), config update (1)
- Audit entries use descriptive action names: `tenant.created`,
  `module.enabled`, `flag.toggled`, `config.updated`
- Platform migration 005 correctly omits FK on `user_id` (platform admins
  live in tenant DBs, not platform DB)
- Sound decision to skip agent/feedback/traces (they use `execution_traces`)

### C2 -- Traefik config cleanup
- Both dev.yml and prod.yml now route all 10 services correctly
- Port mappings are consistent with service code: agent:8004, search:8010,
  traces:8009, feedback:8008
- Prod config applies correct middleware chains: `strip-spoofed-headers` +
  `tenant-from-subdomain` for tenant-scoped routes, omitted for auth/platform
- Platform restricted to `Host(platform.sda.app)` in prod (correct)
- docker-compose.prod.yml defines agent, search, traces with proper
  Dockerfiles, healthchecks, networks, and Traefik labels
- No stale `rag` references in Traefik configs or prod compose

### C3 -- sqlc config consolidation
- All 6 sqlc.yaml files point to centralized paths:
  - Tenant (auth, chat, feedback, ingest, notification): `../../../db/tenant/migrations/`
  - Platform: `../../../db/platform/migrations/`
- Regenerated models include full schema (platform models.go shows new
  `AuditLog` struct alongside existing structs)
- Local `services/*/db/migrations/` directories removed (no stale files)
- Package conventions preserved: platform uses `db` package in-dir, tenant
  services use `repository` in `internal/repository`

### C4 -- SSL mode enforcement
- `ensureSSLMode()` implementation follows the plan spec precisely
- Uses `url.Parse` for safe URL manipulation (no string concatenation)
- Correctly placed in `createPoolLocked()` -- the single point of tenant
  pool creation
- Non-destructive: respects explicit `sslmode=disable` in dev URLs
- Secrets README documents the requirement and the `verify-full` upgrade
  path for when PostgreSQL moves off-host

### C5 -- NATS per-service auth
- `nats-server.conf` defines all 10 users with correctly scoped permissions
- Principle of least privilege applied well:
  - Publish-only services (auth, chat, agent, platform) have empty subscribe
  - Subscribe-only services (notification, traces) have empty publish
  - Bidirectional services (ingest, extractor, feedback) have matching
    publish/subscribe on their own namespace
  - WS Hub has `tenant.>` subscribe (needs to bridge everything to browsers)
- docker-compose.prod.yml uses per-service NATS URLs with embedded
  credentials for 8 of 10 services
- Dev compose maintains no-auth NATS (correct per plan)
- JetStream config preserved with reasonable defaults (128MB mem, 1GB file)

---

## Summary

The PR implements the 5 critical fixes from Plan 08 Phase 1 with solid
core work. Audit logging, Traefik routing, sqlc consolidation, and SSL
enforcement are all well-executed.

Three blockers prevent merge:
1. **B1:** NATS container gets no env vars -> all passwords are empty
2. **B2:** Platform service has no NATS_URL -> lifecycle events dead
3. **B3:** Platform event type dots create 5-segment subjects that NATS
   `tenant.*.notify.*` permission will deny

Five items need correction (D1-D5): stale dev compose `rag` service,
orphan `nats_token` secret, incomplete README, missing DB constraint,
silent parse failure.

After fixing B1-B3 and D1-D5, the PR is ready to merge.
