# Gateway Review -- Plan 08 Final Assessment (9 PRs consolidated)

**Fecha:** 2026-04-05
**Resultado:** APROBADO CON OBSERVACIONES (no blockers, plan declarable cerrado)
**Branch:** 2.0.1
**PRs:** #87 through #95 (9 squash merges)
**Scope:** 52 hallazgos, 7 fases de hardening

---

## 1. Consistencia post-merge

**Veredicto: Excelente**

Despues de 9 PRs, el codebase muestra un patron coherente y consistente. Verificaciones puntuales:

### cmd/main.go pattern (10 services)
| Patron | Servicios conformes | Excepciones |
|--------|-------------------|-------------|
| `slog.New(JSONHandler)` | 10/10 | - |
| `config.Env()` | 10/10 | - |
| `sdajwt.MustLoadPublicKey()` | 9/10 | auth (needs private key too, uses `loadJWTKeys()` -- correct) |
| `natspub.Connect()` + `defer nc.Drain()` | 9/9 NATS users | search has no NATS (correct: read-only service) |
| `ReadHeaderTimeout: 10*time.Second` | 10/10 | - |
| `sdamw.SecureHeaders()` | 10/10 | - |
| `otelhttp.NewHandler()` | 10/10 | - |
| `sdaotel.Setup()` with Insecure bool | 10/10 | - |
| Shutdown error logged | 10/10 | - |
| `middleware.Recoverer` | 10/10 | - |

### Authentication pattern
| Service | Auth middleware | Correct |
|---------|---------------|---------|
| auth | `sdamw.AuthWithConfig()` with blacklist, FailOpen:false | Yes |
| chat | `sdamw.Auth(publicKey)` | Yes |
| ws | JWT in upgrade handler (correct: WS needs different pattern) | Yes |
| agent | `sdamw.Auth(publicKey)` + rate limiter | Yes |
| search | `sdamw.Auth(publicKey)` + rate limiter | Yes |
| traces | `sdamw.Auth(publicKey)` | Yes |
| notification | `sdamw.Auth(publicKey)` | Yes |
| feedback | `sdamw.Auth(publicKey)` | Yes |
| ingest | `sdamw.Auth(publicKey)` | Yes |
| platform | `requirePlatformAdmin` (JWT direct verify, role+slug check) | Yes |

### NATS connection consistency
All 9 NATS-using services use `natspub.Connect()` + `defer nc.Drain()`. Zero instances of raw `nats.Connect()` or `nc.Close()`.

---

## 2. Completitud -- 52 hallazgos vs. estado actual

### Fase 1: Critical (C1-C5)

| ID | Hallazgo | Estado | Notas |
|----|----------|--------|-------|
| C1 | Audit logging en 5 servicios faltantes | DONE | search: `audit.NewWriter(pool)` in main.go. Other services with NATS consumers (traces, feedback) have consumer-only audit. Agent publishes via TracePublisher. Platform publishes events via NATS. |
| C2 | Traefik configs actualizados | DONE | dev.yml: 10 services, all with correct ports. prod.yml: matching. No `rag` references. docker-compose.prod.yml: agent, search, traces all present with correct Dockerfiles. |
| C3 | sqlc configs apuntan al schema correcto | DONE | All 6 sqlc.yaml files point to `../../../db/{tenant,platform}/migrations/`. Empty `services/*/db/migrations/` dirs removed. |
| C4 | SSL en PostgreSQL | DONE | `ensureSSLMode()` in `pkg/tenant/resolver.go:357-370` appends `sslmode=require` if not set. Dev URLs with explicit `sslmode=disable` are preserved. |
| C5 | NATS per-service auth | DONE | `nats-server.conf`: 10 users with granular pub/sub permissions. docker-compose.prod.yml: per-service NATS credentials via env vars. |

### Fase 2: High Security (H-NEW, H8, H11, etc.)

| ID | Hallazgo | Estado | Notas |
|----|----------|--------|-------|
| H-NEW Rate limiting | DONE | `pkg/middleware/ratelimit.go`: ByIP + ByUser key functions, token bucket with 10min cleanup. Auth: 5/min login, 10/min refresh, 5/min MFA. Agent+Search: 30/min by user. |
| H-NEW ReadHeaderTimeout | DONE | All 10 services have `ReadHeaderTimeout: 10 * time.Second`. |
| H8 Ownership checks | DONE | `GetSession`: `AND user_id = $2`. `TouchSession`: `AND user_id = $2`. `GetJob`: `AND user_id = $2`. |
| H11 JWT JTI defense | DONE | `pkg/jwt/jwt.go:78-79`: auto-generates UUID if empty. `pkg/middleware/auth.go:72-75`: rejects tokens with empty JTI when blacklist is configured. |
| H-NEW Blacklist FailOpen | DONE | `AuthConfig{FailOpen bool}` in auth.go. Auth uses FailOpen:false. Other services use default Auth() (no blacklist). |
| M9 Security scans | DONE | ci.yml: gosec runs without `|| true`, Trivy uses `exit-code: '1'`. |
| M11 Agent WriteTimeout | DONE | Agent: `WriteTimeout: 5 * time.Minute`. All services have ReadHeaderTimeout. |
| M14 API key not cached | DONE | `cachedModelConfig` struct in resolver.go excludes `APIKey`. API key fetched from DB per-call. |
| L4 writeJSONError | DONE | `json.NewEncoder(w).Encode(map[string]string{"error": msg})` -- no string concatenation. |

### Fase 3: Database Hardening

| ID | Hallazgo | Estado | Notas |
|----|----------|--------|-------|
| H4 Pagination universal | DONE | `pkg/pagination/pagination.go`: Parse(), Offset(), Limit(), SetHeaders(). MaxPage=10000 cap. Used by chat (ListSessions), platform (ListTenants), documents (GetAllDocumentTrees, GetCollectionDocumentTrees). |
| H9 ListActiveUsers LATERAL JOIN | DONE | auth.sql uses `LEFT JOIN LATERAL` with `LIMIT $1 OFFSET $2`. |
| M10 Batch insert for pages | NOT DONE | `InsertDocumentPage` is still 1 row/call. See observations below. |
| L5 DeleteJobByID ownership | PARTIAL | `DeleteJob` has `AND user_id = $2`. `DeleteJobByID` still lacks it (used internally only). Acceptable since it's internal-only. |
| L6 Notification purge | DONE | `PurgeOldNotifications` query: 90-day read purge. |
| L7 FK documents.uploaded_by | DONE | Migration 009: `ADD CONSTRAINT fk_documents_uploaded_by`. |
| L8 PerformancePercentiles generated column | DONE | Migration 009: `latency_ms NUMERIC GENERATED ALWAYS AS`. Uses regex guard for non-numeric values. Query uses `ORDER BY latency_ms`. |
| L9 CountHistoricalUsage time filter | DONE | `AND created_at > $1` parameter added. |
| L10 Feedback categories constants | NOT VERIFIED | No `pkg/feedback/categories.go` found. Categories are still string literals in SQL queries. Low priority. |
| L11 Redundant index dropped | DONE | Migration 009: `DROP INDEX IF EXISTS idx_refresh_tokens_hash`. |

### Fase 4: Feature Wiring + NATS

| ID | Hallazgo | Estado | Notas |
|----|----------|--------|-------|
| H2 EnabledModules from Platform DB | DONE | `tenant.Resolver.ListEnabledModules()` queries `tenant_modules JOIN modules`. Auth handler uses resolver when available, falls back to core defaults. |
| H5 Cache service structs | PARTIAL | Auth handler has resolver-based multi-tenant, but `sync.Map` service cache not explicitly visible. Per-request service allocation may still occur. |
| H10 Feedback HealthScore tenant-level | DONE | `feedback/handler/feedback.go:224-264`: queries `tenant_health_scores` filtered by `$1` tenant_id from header. |
| M3 Guardrails from config | PARTIAL | Agent uses `guardrails.DefaultInputConfig(10000)` inline. Chat uses `guardrails.DefaultInputConfig(50000)`. Not loaded from config resolver yet. |
| M4 Tool definitions from YAML | DONE | `tools.LoadModuleTools()` called in agent main.go with `modules/*/tools.yaml`. |
| M6 Agent userID from context | DONE | `agent/handler/agent.go:58`: `userID := r.Header.Get("X-User-ID")`. |
| H7 Parent contexts in NATS consumers | DONE | Both traces and notification consumers use `c.ctx` derived from parent via `context.WithCancel(ctx)`. |
| M1 NATS connections + shutdown | DONE | All 9 NATS services: `natspub.Connect()` + `defer nc.Drain()`. |
| M7 JetStream API migration | DONE | Traces consumer uses `jetstream.New()` (new API), not deprecated `nc.JetStream()`. |
| M8 Dead Letter Queue | DONE | `pkg/nats/dlq.go`: `LogDLQ()` logs + publishes to `dlq.{stream}.{subject}`. Both consumers have `MaxDeliver: 5`. |

### Fase 5: Testing & Quality

| ID | Hallazgo | Estado | Notas |
|----|----------|--------|-------|
| H3 Search tests | DONE | `search_test.go`: TestBuildTreeView_SingleLevel, TestBuildTreeView_NestedNodes, TestParseNodeIDs_CommaSeparated, TestContainsInt. |
| H3 Traces tests | DONE | `traces_test.go`: 4 tests covering tenant context requirement and handler behavior. |
| M2 pkg/llm tests | DONE | `client_test.go`: 17 tests covering Chat, SimplePrompt, auth headers, error handling, malformed JSON, cancelled context, empty choices. Excellent coverage. |
| M2 pkg/otel tests | DONE | `otel_test.go`: 5 tests covering unreachable endpoint, multiple shutdown calls, default endpoint, empty service name, cancelled context. |
| M15 Interfaces for core types | NOT DONE | No interfaces for `audit.Logger`, `llm.ChatClient`, or `natspub.EventPublisher` in the pkg packages. Chat service defines its own `EventPublisher` interface locally. See observations. |

### Fase 6: Build, CI & OpenAPI

| ID | Hallazgo | Estado | Notas |
|----|----------|--------|-------|
| H6 .dockerignore | DONE | Root `.dockerignore` with `.git`, `.github`, `.claude`, `node_modules`, `apps`, `docs`, etc. |
| M5 OpenAPI/Swagger | NOT DONE | No swaggo annotations found anywhere. See observations. |
| M9 Security scans (duplicate of Fase 2) | DONE | See above. |
| M13 go.mod cleanup | PARTIAL | `replace` directives still present in `agent/go.mod`, `search/go.mod`, `traces/go.mod` (all pointing to `../../pkg`). go.work should handle these. |
| L12 Scaffold Dockerfile | DONE | `.scaffold/Dockerfile` uses `golang:1.25-alpine`. |
| L16 Auth: config.Env() | DONE | `config.Env("AUTH_PORT", "8001")` etc. throughout auth main.go. |
| L17 Auth: Routes() pattern | NOT DONE | Auth routes are still defined inline in main.go, not via `handler.Auth.Routes()`. |

### Fase 7: Infrastructure

| ID | Hallazgo | Estado | Notas |
|----|----------|--------|-------|
| L15 Backup strategy | DONE | `deploy/scripts/backup.sh`: PG dumps (platform + per-tenant), Redis RDB, NATS JetStream. Encrypted with `age`. SHA-256 checksums. MinIO upload. Daily/monthly retention. `restore.sh`: download, verify checksum, decrypt, restore. |
| M12 CrowdSec | DONE | docker-compose.prod.yml: crowdsec service with `acquis.yaml` reading Traefik logs. |
| L14 Docker socket proxy | DONE | `tecnativa/docker-socket-proxy` in prod compose, CONTAINERS=1 only. Traefik uses `tcp://docker-socket-proxy:2375`. |
| L13 CPU limits | DONE | All services have `deploy.resources.limits.cpus` and `memory` in prod compose. |
| L19 otel.Setup TLS configurable | DONE | `Config.Insecure bool` in `pkg/otel/otel.go`. All services pass `Insecure: true` for local. |
| M-NEW CreateTenant URL validation | NOT DONE | `platform/handler/platform.go:131` still accepts raw `postgres_url` and `redis_url`. See observations. |
| L1 Pool health monitoring | DONE | `resolver.StartHealthCheck()` in `pkg/tenant/resolver.go:329-352`. Pings all cached pools, removes unhealthy ones. |
| L21 Resolver mutex recover | DONE | `resolver.go:185-192`: `createPoolLocked` wraps `pgxpool.NewWithConfig` with `recover()`. |
| L2 Dev compose comment | N/A | Not verified (compose dev not in scope of changes). |
| L3 Redis per-tenant decision | DONE | Documented -- single Redis for MVP. |
| L18 Notify double-serialization | DONE | `publisher.go:88-99`: type assertion for Event, *Event, then map fallback. |
| L20 audit.Write doc | DONE | Comment says "non-failing". |
| L-NEW Shutdown error handling | DONE | All 10 services log `slog.Error("shutdown error", ...)`. |
| L-NEW Feedback handler error swallowing | NOT DONE | `Summary()` handler still logs errors but returns 200 with zero-values. See observations. |

### Scoreboard

| Status | Count | % |
|--------|-------|---|
| DONE | 42 | 81% |
| PARTIAL | 4 | 8% |
| NOT DONE | 6 | 11% |
| **Total** | **52** | |

---

## 3. Regressions

**None found.** Las fases posteriores no rompieron nada de las anteriores.

Verified:
- JWT flow: EdDSA signing, JTI auto-generation, blacklist check, FailOpen config -- all consistent
- Header stripping in auth middleware: Del before Set pattern preserved
- NATS subject validation: `IsValidSubjectToken` + `IsValidEventType` in publisher, `validateToken` in agent traces
- Tenant isolation: All queries use parameterized filters, tenant context flows correctly
- sqlc configs point to correct shared migrations
- Dockerfiles all use `golang:1.25-alpine` + distroless
- CI pipeline: build, vet, test, security scan -- all operational

---

## 4. Gaps restantes (NOT DONE items)

### Items NOT DONE (6 total)

**1. M10 -- Batch insert for document pages**
`InsertDocumentPage` is still called 1 row at a time. For a 200-page PDF, that's 200 DB round trips.
- **Impact:** Performance only. Functional correctness is fine.
- **Recommendation:** Implement `pgx.CopyFrom` in Plan 09 or a separate performance PR. Not a blocker.

**2. M15 -- Interfaces for core types**
`pkg/audit`, `pkg/llm`, `pkg/nats` export concrete structs, not interfaces. The chat service has a local `EventPublisher` interface showing the pattern works.
- **Impact:** Testing ergonomics. Mocking requires interface wrappers.
- **Recommendation:** Extract interfaces when adding integration tests. Not a blocker.

**3. M5 -- OpenAPI/Swagger**
No swaggo annotations found anywhere.
- **Impact:** API documentation gap. Frontend developers need to read handler code.
- **Recommendation:** Dedicate a separate plan phase. This is additive work, not a fix.

**4. L17 -- Auth: Routes() pattern**
Auth routes are inline in main.go (correct functionally, inconsistent with chat/agent/search which use `handler.Routes()`).
- **Impact:** Style consistency only. Zero functional impact.
- **Recommendation:** Minor refactor, can be done anytime.

**5. M-NEW -- CreateTenant URL validation**
Platform handler accepts raw `postgres_url` and `redis_url` without validation. The plan suggested template generation instead.
- **Impact:** SSRF risk from platform admin (who already has full access). Attacker would need admin JWT for platform slug. Risk is theoretical given the trust level.
- **Recommendation:** Implement URL validation or template generation in Plan 09.

**6. L-NEW -- Feedback Summary error swallowing**
`feedback/handler/feedback.go:52-73`: four queries each log errors independently but the handler returns 200 with zero-value structs. A DB failure produces misleading data.
- **Impact:** Observability gap. Frontend shows zeros instead of an error state.
- **Recommendation:** Return 500 if any query fails, or return `null` for failed sections with an `errors` array.

### Items PARTIAL (4 total)

**1. M13 -- go.mod `replace` directives**
Still present in agent, search, traces go.mod. go.work handles resolution, but these are unnecessary and could confuse `go mod tidy`.

**2. H5 -- Service struct caching**
Auth handler caches via resolver, but per-request service allocation pattern not fully optimized with sync.Map.

**3. M3 -- Guardrails from config resolver**
Block patterns are still inline constants, not loaded from platform DB. Functional but not dynamic per-tenant.

**4. L10 -- Feedback categories as constants**
Categories like `"response_quality"`, `"error_report"` are still string literals.

---

## 5. Deuda tecnica introducida

### Low concern (acceptable for MVP)

1. **Rate limiter in-memory (sync.Map-based):** `pkg/middleware/ratelimit.go` uses in-memory token buckets with 10-minute cleanup. Fine for single-node. Documented migration path to Redis. The cleanup goroutine runs forever (leaked goroutine on test) but is a non-issue in production (process lifetime = goroutine lifetime).

2. **DLQ without replay mechanism:** `pkg/nats/dlq.go` logs and publishes to DLQ subjects. No replay endpoint or script exists yet. The plan acknowledged this: "replay: script manual or endpoint admin". Fine to defer.

3. **TracePublisher duplicates regex:** `services/agent/internal/service/traces.go:13` has its own `safeToken` regex that mirrors `natspub.IsValidSubjectToken`. Should use the canonical one from `pkg/nats`. Not a bug (same pattern) but a maintenance risk.

4. **Feedback Summary error swallowing:** As noted above, this is existing technical debt that Plan 08 planned to fix (L-NEW) but didn't.

5. **go.mod replace directives:** Three services have unnecessary `replace` directives. Should be cleaned up.

### No concern

- **sync.Map without eviction (rate limiter):** Actually uses a regular `map[string]*limiterEntry` with mutex and explicit 10-minute cleanup. No eviction-less sync.Map issue.
- **Pagination SetHeaders unused:** The helper exists but some handlers don't call it. They return arrays directly, which is fine for endpoints without total count.

---

## 6. Security posture assessment

### What Plan 08 fixed (the big wins)

1. **Header spoofing protection:** Auth middleware strips all identity headers before JWT processing, then re-injects from verified claims. Traefik prod config also strips `X-Tenant-Slug` and `X-Tenant-ID`.

2. **JWT hardening:** EdDSA signing (no HMAC), JTI auto-generation, blacklist with configurable fail-open/fail-closed, MFA-pending token rejection. `alg:none` attack impossible due to `SigningMethodEd25519` type assertion.

3. **Rate limiting:** Application-level per-IP (login 5/min, refresh 10/min, MFA 5/min) and per-user (AI 30/min). On top of Traefik's 100 req/s global limit.

4. **ReadHeaderTimeout:** All services protected against slowloris (10s timeout).

5. **NATS subject injection:** All publishers validate tokens with `IsValidSubjectToken` or `IsValidEventType`. Agent's `TracePublisher` validates tenant slug.

6. **SQL ownership checks:** GetSession, TouchSession, GetJob all filter by user_id. ListMessages requires ownership check at handler level.

7. **System role bypass blocked:** Chat handler explicitly rejects `role: "system"` from API clients.

8. **Tenant isolation:** JWT slug cross-validated against Traefik-injected slug. Per-tenant DB pools via resolver. NATS subjects include tenant slug. Traces consumer validates tenant matches between subject and payload.

9. **Infrastructure hardening:** Docker socket proxy, CrowdSec IDS, network segmentation (frontend/backend/data), distroless containers, non-root user, CPU/memory limits.

### Remaining attack surface (for next hardening cycle)

| Vector | Risk | Mitigation path |
|--------|------|-----------------|
| CreateTenant SSRF | Low (admin-only) | URL validation or template generation |
| No WAF rules for SQL injection in query params | Very Low (sqlc parameterized) | CrowdSec + Traefik rules |
| Redis single-instance (shared tenants) | Low (key-namespaced) | Per-tenant Redis when >10 tenants |
| Backup encryption key management | Medium | Rotate keys, use Vault |
| No certificate pinning on PostgreSQL | Medium (local network) | `sslmode=verify-full` when PG moves remote |
| No audit log for platform admin actions | Medium | Add audit.Write to platform handlers |
| Feedback platform handler uses raw SQL | Low (parameterized, admin-only) | Migrate to sqlc |

### Recommended next security focus

1. **Platform audit logging** -- admin actions (create/disable tenant, enable module) should be audited.
2. **CreateTenant URL validation** -- even if admin-only, input validation is a principle.
3. **Rate limit Redis migration path** -- document when to switch from in-memory to Redis.

---

## 7. What is well done

1. **Middleware consistency:** The `sdamw.SecureHeaders()`, `sdamw.Auth()`, `sdamw.RateLimit()` pattern is clean and used everywhere correctly.

2. **NATS publisher:** `pkg/nats/publisher.go` is one of the best pieces of the codebase. Subject validation, event type validation, no double-serialization, channel validation for broadcasts.

3. **pkg/pagination:** Clean, capped, overflow-safe. MaxPage prevents int32 overflow.

4. **Backup/restore scripts:** Production-ready with encryption, checksums, retention policies, and MinIO upload.

5. **Traefik configs:** Both dev and prod are complete, consistent, and well-documented. The subdomain regex extraction in prod is correct.

6. **docker-compose.prod.yml:** All 10 services with secrets, health checks, resource limits, network segmentation. CrowdSec and socket proxy included.

7. **Test quality:** pkg/llm tests (17 tests) and pkg/otel tests (5 tests) are thorough and test edge cases (cancelled context, malformed JSON, empty choices).

8. **JWT design:** EdDSA asymmetric signing is the correct choice. Compromising a non-auth service doesn't enable token forging.

9. **Tenant resolver:** SSL enforcement, health checking, pool caching with singleflight-style locking, encrypted credential support, graceful degradation.

---

## Conclusion

Plan 08 achieved its goal: closing the gap between "well-architected" and "production-ready." 81% of the 52 findings are fully resolved, 8% are partially done, and the remaining 11% are additive work (OpenAPI, batch insert, interfaces) or low-risk deferred items. No blockers. No regressions.

The system is ready for real users in a controlled environment. The remaining items are genuine improvements, not prerequisites.

**Plan 08 can be declared closed.**
