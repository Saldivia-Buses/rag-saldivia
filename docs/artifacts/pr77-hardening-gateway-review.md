# Gateway Review -- PR #77 Hardening Fixes

**Fecha:** 2026-04-05
**Resultado:** CAMBIOS REQUERIDOS

**Scope:** Security + observability hardening across agent, search, traces, extractor services. Targeted fixes for audit findings C1, C2, H1, H2, H4, H5, O1, O3, O4, O7.

---

## Bloqueantes

### B1. C2 tenant validation only applied to `traces.start`, not `traces.end` or `traces.event`

**Archivo:** `services/traces/cmd/main.go:119-147`

The `extractSubjectTenant()` + mismatch check (lines 107-112) is only in the `traces.start` subscriber. The `traces.end` and `traces.event` subscribers have no C2 validation at all.

While `TraceEndEvent` and `TraceEvent` structs currently lack a `TenantID` field (so you can't do a payload-vs-subject comparison), the subject itself still carries the tenant slug. A compromised publisher can publish to `tenant.victim.traces.end` with a forged `trace_id`, updating another tenant's trace via `RecordTraceEnd` which does `UPDATE ... WHERE id = $9` with no `tenant_id` filter.

**Fix:** Two options:

1. **(Minimal)** Add `TenantID string` to `TraceEndEvent` and `TraceEvent`, then apply the same C2 validation to all three subscribers. Also add `AND tenant_id = $10` to the `RecordTraceEnd` UPDATE WHERE clause.
2. **(Lighter)** Extract the tenant slug from the NATS subject in `traces.end` and `traces.event` subscribers too, and pass it to the service layer so the SQL queries can filter by it.

Option 1 is more robust and consistent with the C2 fix intent.

---

## Debe corregirse

### D1. Search service SQL queries have no `tenant_id` filtering -- C1 fix is incomplete

**Archivos:** `services/search/internal/handler/search.go:44-48`, `services/search/internal/service/search.go:127-140, 281-284`

The C1 fix correctly reads `tenant.FromContext()` in the handler (line 44) and returns 401 if missing. But the tenant info is never passed to `SearchDocuments()` -- the service method signature is `SearchDocuments(ctx, query, collectionID, maxNodes)` with no tenant parameter.

The SQL queries (`loadTrees`, `extractPages`) query `document_trees`, `documents`, `document_pages` tables with **zero `tenant_id` filtering**:

```sql
-- loadTrees without collection filter:
SELECT dt.document_id, d.name, dt.doc_description, dt.tree
FROM document_trees dt JOIN documents d ON d.id = dt.document_id
WHERE d.status = 'ready'

-- extractPages:
SELECT page_number, text, tables, images FROM document_pages
WHERE document_id = $1 AND page_number >= $2 AND page_number <= $3
```

The current architecture uses a single `POSTGRES_TENANT_URL` pool, so if this eventually connects to a shared DB or if the tenant resolver returns the wrong pool, there's no SQL-level defense.

**Fix:** Either:
- (a) Add `tenant_id` column to `documents` / `document_trees` / `document_pages` and filter all queries, or
- (b) Document that tenant isolation relies entirely on per-tenant DB connections and add a TODO for defense-in-depth

At minimum, pass `ti.Slug` or `ti.ID` to the service layer so it's available when the schema adds `tenant_id`.

### D2. Search service has no `RequirePermission` middleware

**Archivo:** `services/search/internal/handler/search.go:29`

The handler routes have no RBAC middleware:
```go
r.Post("/query", h.SearchDocuments)
```

The agent service correctly applies `RequirePermission("chat.read")` (H5 fix), but the search endpoint that agents call is unprotected. Any authenticated user can search all documents regardless of their permissions. This should be at least `collections.read` (matching the RAG service pattern from Plan 04).

### D3. `OTEL_EXPORTER_OTLP_ENDPOINT` in `x-env-common` points to `otel-collector:4317` which is not in docker-compose.dev.yml

**Archivo:** `deploy/docker-compose.dev.yml:37`

The env var `OTEL_EXPORTER_OTLP_ENDPOINT: otel-collector:4317` references a service that lives in a separate compose file (`deploy/observability/docker-compose.observability.yml`). When running with `make dev` (infra only), Go services on the host can't resolve `otel-collector`. When running with `--profile full`, containerized services won't find it either unless the observability stack is up on the same Docker network.

The `pkg/otel.Setup()` handles this gracefully (returns error, caller logs warning and continues), so it won't crash. But the O4 fix description implies this is functional, which it isn't without the observability stack running.

**Fix:** Add a comment noting the dependency on the observability stack, or set the fallback to `localhost:4317` which the `Setup()` function already defaults to.

---

## Sugerencias

### S1. `RecordTraceEnd` UPDATE query lacks `tenant_id` in WHERE clause

`services/traces/internal/service/traces.go:69-79` -- `UPDATE execution_traces SET ... WHERE id = $9` should ideally be `WHERE id = $9 AND tenant_id = $10` for defense-in-depth. Since this is a platform DB (not per-tenant), every tenant's traces coexist. Without the tenant filter, a forged trace ID from NATS could overwrite another tenant's trace record. Low priority since NATS is internal, but aligns with the "security is a constraint" principle.

### S2. Search service `httpClient` does not use `otelhttp.NewTransport`

`services/search/internal/service/search.go:63` -- The `httpClient` for LLM calls uses a plain `http.Client{Timeout: 60*time.Second}`. The O3 fix applied `otelhttp.NewTransport` to the agent's LLM adapter and tool executor, but the search service's own LLM client was missed. Trace propagation to the LLM endpoint won't work for search queries.

### S3. Extractor `ack_wait=300` should be `300e9` (nanoseconds) or use float seconds

`services/extractor/main.py:146` -- The `ConsumerConfig(ack_wait=300)` is in seconds for nats-py, but verify the unit. In nats-py >= 2.9, `ack_wait` is in nanoseconds when passed as an integer to the JetStream consumer config. If this means 300 nanoseconds instead of 300 seconds, messages will be redelivered almost immediately. The comment says "5 min for large PDFs" which would be 300 seconds, but confirm the nats-py API expects seconds or nanoseconds here.

### S4. Consider structured logging in extractor health handler

`services/extractor/main.py:65` -- The `_HealthHandler` suppresses all access logs (`log_message` is a no-op), which is correct for avoiding noise, but consider logging startup once via the structured JSON logger for consistency.

---

## Lo que esta bien

1. **C1 search handler tenant check** -- Correctly reads `tenant.FromContext()` before processing, returns 401 on missing tenant. Pattern matches other services (traces, chat).

2. **C2 `extractSubjectTenant` for traces.start** -- Subject parsing and payload mismatch check is clean. The `parts[1]` extraction from `tenant.{slug}.traces.{action}` is correct.

3. **H1 `MaxReconnects(-1)`** -- Traces service NATS connection now reconnects indefinitely. This prevents silent disconnection after default 60 attempts.

4. **H2 `natsCtx` factory function** -- The pattern of returning `(context.Context, context.CancelFunc)` from a factory and deferring cancel in each subscriber callback is clean. Eliminates the goroutine leak from the previous single-context approach.

5. **H4 storage key tenant prefix validation** -- `pipeline.py:54-58` validates `job.storage_key.startswith(f"{job.tenant_slug}/")` before downloading. Prevents cross-tenant MinIO access. Simple and effective.

6. **H5 `RequirePermission("chat.read")` on agent routes** -- Both `/query` and `/confirm` are protected. Admin bypass works correctly in the middleware.

7. **O1 `sdaotel.Setup()` across all three services** -- Consistent pattern: setup with service name/version, defer shutdown, warn-and-continue on failure. The `pkg/otel` package is well-structured with batch processing and proper resource attributes.

8. **O3 `otelhttp.NewTransport` in agent LLM adapter and tool executor** -- Trace context propagation to downstream services and LLM endpoints. Both `llm.Adapter` and `tools.Executor` use instrumented transports.

9. **O4 `OTEL_EXPORTER_OTLP_ENDPOINT` in env-common** -- All containerized services get the OTel collector endpoint. Override works for host-mode development.

10. **O7 `python-json-logger` in extractor** -- Structured JSON logging with field renaming matches the Go services' slog JSON output format. Clean integration with `logging.basicConfig`.

11. **Ed25519 JWT throughout** -- All reviewed services use `loadPublicKey()` with `ParsePublicKeyEnv()`, verifying EdDSA tokens. No lingering HS256 code.

12. **NATS subject injection validation in extractor** -- `_validate_subject_token()` with `_SAFE_SUBJECT_RE` rejects dots, wildcards, whitespace in tenant_slug and document_id before constructing NATS subjects. This was a blocker in the PR #66 review and is now fixed.

13. **Auth middleware header stripping** -- `pkg/middleware/auth.go` strips `X-User-ID`, `X-User-Email`, `X-User-Role`, `X-Tenant-ID`, `X-Tenant-Slug` before processing JWT. Prevents spoofing.

14. **Security headers middleware** -- Applied consistently across all three services with correct values (nosniff, DENY frame, strict referrer, permissions policy, conditional HSTS).

15. **`MaxBytesReader` on all handlers** -- Search (64KB), agent query (256KB), agent confirm (64KB). Prevents memory exhaustion from oversized payloads.
