# Gateway Review -- PR #105 Astro Service Scaffold (Phase 1, Plan 11)

**Fecha:** 2026-04-06
**Resultado:** CAMBIOS REQUERIDOS

---

## Bloqueantes

### B1. `FailOpen: false` rejects requests when Redis is unreachable [CRITICAL]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/cmd/main.go:93`

Astro is the **only non-auth service** that uses `FailOpen: false`. Every other tenant-scoped service (chat, search, agent, notification, traces, feedback, ingest) uses `FailOpen: true`. With `FailOpen: false`, if Redis is temporarily unreachable, **all authenticated requests get 503**. This is correct for auth (login gate) but wrong for a data service where availability is more important than immediate revocation enforcement.

```go
// Current (incorrect for a data service):
FailOpen: false,

// Fix:
FailOpen: true,
```

**Rationale:** `FailOpen: true` means "if I can't check the blacklist, let the JWT through." The token is still cryptographically verified. The blacklist is a secondary check for explicit revocation (logout/password change). Blocking all traffic because Redis blipped is a worse outcome for a non-critical path.

---

### B2. `middleware.Timeout(5 * time.Minute)` applies to SSE `/v1/astro/query` -- will kill long streams [CRITICAL]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/cmd/main.go:82`

Chi's `middleware.Timeout` wraps the handler in a context with deadline. The SSE `Query` endpoint streams events and will be **canceled after 5 minutes** even if the LLM is still generating. The agent service uses 90s timeout globally but sets `WriteTimeout: 5 * time.Minute` on the HTTP server -- a different strategy.

**Fix:** Either:
- (a) Move the SSE route to a separate `r.Group` that **omits** the Timeout middleware (like WS Hub does), or
- (b) Set `middleware.Timeout(0)` for the SSE group (which disables the context deadline).

The HTTP server's `WriteTimeout: 5 * time.Minute` already protects against indefinite slowloris. The chi middleware timeout is the problem, not the server timeout.

---

### B3. Missing `user_id` column in contacts migration [CRITICAL]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/db/migrations/001_contacts.sql`

The table has `tenant_id` but no `user_id`. In SDA, every user-created resource is owned by the user who created it. Without `user_id`:
- Any user in the tenant can see/edit all contacts (no row-level access control)
- No audit trail of who created what
- Other services (notification, chat, ingest, feedback) all track `user_id` on their entities

**Fix:** Add `user_id UUID NOT NULL` and update the unique constraint and/or index:

```sql
user_id     UUID NOT NULL,
-- ...
CREATE INDEX idx_contacts_tenant_user ON contacts(tenant_id, user_id);
```

---

### B4. Migration missing DOWN section [CRITICAL]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/db/migrations/001_contacts.sql`

Bible rule: "Migrations tienen UP y DOWN." The migration only has the CREATE and INDEX statements but no `-- +goose Down` (or equivalent) section with `DROP TABLE IF EXISTS contacts;`. Without DOWN, rollbacks are impossible.

**Fix:** Add a down section (format depends on migration tool -- likely goose or golang-migrate):

```sql
-- +goose Down
DROP TABLE IF EXISTS contacts;
```

---

## Debe corregirse

### M1. `tenant_id TEXT` should be `UUID` [MEDIUM]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/db/migrations/001_contacts.sql:3`

`tenant_id` is declared as `TEXT` but tenant IDs are UUIDs everywhere in SDA (the Resolver returns UUID strings, the JWT claim `tid` is a UUID). Using `TEXT` means no database-level validation that the value is actually a UUID, and JOIN performance against any platform table is worse.

**Fix:** `tenant_id UUID NOT NULL`

---

### M2. Spanish column name `tipo` breaks conventions [MEDIUM]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/db/migrations/001_contacts.sql:17`

Bible: "Idioma codigo: ingles." The column `tipo` should be `type` or `contact_type` (to avoid the SQL reserved word `type`).

```sql
-- Fix:
contact_type TEXT NOT NULL DEFAULT 'person',
```

---

### M3. Dockerfile uses `alpine` runtime instead of `distroless` [MEDIUM]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/Dockerfile:16`

Every other SDA service uses `gcr.io/distroless/static-debian12` as the runtime image. Astro uses `alpine:3.22`. This is understandable because Phase 2 will need CGO (swephgo requires C libs), but:

1. The deviation should have a comment explaining **why** (CGO runtime dependency on musl/glibc).
2. For Phase 1 (stubs only, no actual sweph calls), it could still use distroless since `ephemeris.Init()` and `Close()` are no-ops.

**Fix (Phase 1):** Add a comment: `# CGO required for swephgo (Phase 2) -- requires musl libc at runtime`

Or even better for Phase 1, use distroless now and switch to alpine only when Phase 2 actually needs it.

---

### M4. `POSTGRES_TENANT_URL` is optional -- silent degradation [MEDIUM]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/cmd/main.go:62-69`

When `dbURL` is empty, the pool is nil and the handler gets a nil pool. This means `ListContacts` and `CreateContact` (which need DB) will panic on nil pointer dereference when they're implemented. The chat service **exits** when DB URL is missing. Search also **exits**.

**Fix:** Either make DB required (like chat/search):
```go
if dbURL == "" {
    slog.Error("POSTGRES_TENANT_URL is required")
    os.Exit(1)
}
```

Or, if you genuinely want to support DB-less mode for pure ephemeris calculations, then the handler must nil-check `h.db` and return 503 in contact endpoints. But that complicates things for no real gain in a scaffold.

---

### M5. Missing README.md and CHANGELOG.md [MEDIUM]

**Archivos faltantes:** `services/astro/README.md`, `services/astro/CHANGELOG.md`

Bible: "Servicio nuevo -> Spec + README del servicio + CLAUDE.md" and the service structure template shows both files. The README should document endpoints, port, env vars, and the CGO deviation.

---

### M6. Traefik dev routing not configured for astro [MEDIUM]

**Archivo:** `/Users/enzo/rag-saldivia/deploy/traefik/dynamic/dev.yml`

No router or service entry for `astro` on port 8011. Without this, the service is not reachable through Traefik in dev. Every other service has an entry.

**Fix:** Add to `dev.yml`:
```yaml
routers:
  astro:
    rule: "PathPrefix(`/v1/astro`)"
    service: astro
    entryPoints: [web]
    middlewares: [dev-cors, dev-tenant]

services:
  astro:
    loadBalancer:
      servers:
        - url: "http://host.docker.internal:8011"
```

---

### M7. `sqlc.yaml` queries path points to non-existent file [MEDIUM]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/db/sqlc.yaml:4`

The config references `../internal/repository/queries.sql` but that file does not exist yet. Other services use the pattern `queries/` (a directory). This will fail `make sqlc`.

Also, `schema: "migrations/"` references the local migrations dir, while other tenant-scoped services reference the shared schema at `../../../db/tenant/migrations/`. If astro contacts live in the tenant DB (which they should -- they're tenant-scoped data), the schema source should include both the shared tenant schema AND the astro-specific migration.

---

### M8. No `emit_empty_slices: true` in sqlc config [LOW]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/db/sqlc.yaml`

Chat, notification, and other services use `emit_empty_slices: true` so list endpoints return `[]` instead of `null` in JSON. Astro's config is missing this.

---

## Sugerencias

### S1. Rate limit of 10 req/min is very aggressive [LOW]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/cmd/main.go:96-99`

10 requests per minute per user is the tightest rate limit in the system. For comparison:
- Search: 30/min
- Agent: 30/min

A user computing multiple chart types for a single contact (natal + transits + progressions + returns) would hit the limit in one flow. Consider 30/min to match other AI-backed services, or at minimum 20/min.

---

### S2. `UNIQUE(tenant_id, lower(name))` -- functional unique index would be better [LOW]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/db/migrations/001_contacts.sql:19`

`UNIQUE(tenant_id, lower(name))` in a table constraint is PostgreSQL-specific and works, but the convention is to use a separate unique index for functional expressions:

```sql
CREATE UNIQUE INDEX uq_contacts_tenant_name ON contacts(tenant_id, lower(name));
```

This also makes it easier to add `user_id` to the constraint when B3 is fixed (likely you want unique per user, not per tenant):

```sql
CREATE UNIQUE INDEX uq_contacts_user_name ON contacts(tenant_id, user_id, lower(name));
```

---

### S3. `utc_offset INTEGER` is fragile for DST [INFO]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/db/migrations/001_contacts.sql:12`

UTC offset as integer (-3) doesn't handle daylight saving time transitions. If this is birth data, the offset at birth is fixed and an integer is fine. But if this is also used for transit/progression targets, consider storing IANA timezone name (`America/Buenos_Aires`) and computing offset dynamically. This is a domain decision, not a bug.

---

### S4. Consider `sseError` unused -- dead code [INFO]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/internal/handler/astro.go:73`

`sseError()` is defined but never called. It's presumably there for Phase 2, which is fine for a scaffold, but worth noting.

---

## Lo que esta bien

1. **SDA bootstrap pattern followed correctly.** Logger setup (JSON handler), signal handling, OTel init, graceful shutdown -- all match the established pattern from chat/search/agent.

2. **Auth middleware applied correctly.** Uses `AuthWithConfig` with public key + blacklist, applied as a group middleware to all `/v1/astro/*` routes. Health endpoint is outside the group.

3. **Secure headers middleware present.** `sdamw.SecureHeaders()` applied in the correct order.

4. **JWT via Ed25519 public key.** Uses `MustLoadPublicKey("JWT_PUBLIC_KEY")` matching the new asymmetric pattern.

5. **Token blacklist integrated.** `security.InitBlacklist` connected to shared Redis, passed to auth config.

6. **OTel instrumentation.** Both `sdaotel.Setup()` and `otelhttp.NewHandler()` present.

7. **Server timeouts well-configured.** ReadTimeout, ReadHeaderTimeout, WriteTimeout, IdleTimeout all set explicitly. WriteTimeout of 5min matches the SSE streaming nature.

8. **SSE pattern is correct.** Proper headers (`text/event-stream`, `no-cache`, `keep-alive`, `X-Accel-Buffering: no`), flusher type assertion, event format with `event:` + `data:` + double newline.

9. **Go workspace updated.** `go.work` includes `./services/astro`.

10. **VERSION file present.** `0.1.0` is correct for a scaffold.

11. **Ephemeris abstraction clean.** `ephemeris.Init()` / `Close()` provides a clean seam for Phase 2 without leaking swephgo details.

12. **LLM client uses `pkg/llm.ChatClient` interface.** Correct dependency injection, allows mocking in tests.

13. **Handler stubs are comprehensive.** All 13 endpoints from the plan (natal, transits, solar-arc, directions, progressions, returns, profections, firdaria, fixed-stars, brief, query + contacts CRUD) are present with consistent 501 responses.

14. **Golden test data present.** `testdata/golden/` has reference outputs for major calculation types -- good foundation for Phase 2 testing.
