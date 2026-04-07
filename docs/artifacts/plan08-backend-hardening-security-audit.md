# Security Audit -- Plan 08 Backend Hardening -- 2026-04-05

## Resumen ejecutivo

Plan 08 proposes fixing 52 findings (5 critical, 11 high, 15 medium, 21 low)
across 8 phases. The security fixes are mostly correct but several proposed
solutions have gaps, and the audit missed at least 6 additional vulnerabilities
in the current codebase. The plan cannot ship as-is -- this report details the
corrections needed before each phase is implemented.

**Overall assessment:** The plan's instincts are good. The priority ordering is
mostly correct. But several proposed fixes are incomplete, one severity is
wrong, and the gRPC phase (4) has no security design at all.

---

## 1. Evaluation of Proposed Solutions

### C2 -- Rate Limiting Middleware

**Proposed:** `golang.org/x/time/rate` in-memory or Redis-backed. 3 tiers:
global 100/s per IP, auth 5/min per IP on login, AI 30/min per user on agent/search.

**Assessment: PARTIALLY CORRECT -- gaps.**

Correct:
- 3-tier approach is right
- IP-based for global and auth, user-based for AI is correct
- `429 + Retry-After` is correct

Problems:
1. **Traefik already has rate limiting in prod.** The prod config at
   `deploy/traefik/dynamic/prod.yml:110-113` has `rateLimit: average:100, burst:200`.
   The plan's Go-level middleware would be a second layer. This is fine (defense-in-depth)
   but the plan should state the layering explicitly to avoid confusion.

2. **Missing vector: refresh endpoint.** `/v1/auth/refresh` is not mentioned. A
   compromised refresh token could be used in a tight loop to generate unlimited
   access tokens. Rate limit refresh to 10/min per IP.

3. **Missing vector: MFA verify.** `/v1/auth/mfa/verify` allows TOTP brute force.
   6-digit TOTP has 1M possibilities. At 5 attempts/s unrestricted, that's 200k
   seconds worst case, but with timing windows (30s TOTP window) the effective
   space is smaller. Rate limit to 5/min per IP, same as login.

4. **Missing vector: password reset (future).** When password reset exists, the
   reset endpoint needs the same rate limiting as login. Flag this for when it
   ships.

5. **In-memory vs Redis:** In-memory is fine for single-instance dev. For prod
   (even single node), Redis-backed is required because Traefik restarts lose
   the limiter state. The plan says "or Redis-backed" -- it should say "Redis-backed
   in prod, in-memory acceptable in dev only."

**Severity of gaps: HIGH (MFA brute force) + MEDIUM (refresh flooding).**

---

### C4 -- sslmode

**Proposed:** `sslmode=require` in prod, `sslmode=disable` in dev.

**Assessment: INSUFFICIENT for production.**

- `sslmode=require` encrypts the connection but does NOT verify the server's
  identity. A MITM can present any certificate and the client will accept it.
- **Correct for this architecture:** Since PostgreSQL runs on the same host (inhouse
  workstation), and traffic is localhost/Docker-internal, `sslmode=require` is
  acceptable TODAY. But the plan should document this decision explicitly.
- **If PostgreSQL ever moves off-host:** must upgrade to `sslmode=verify-full` with
  CA cert pinning. The plan should add a TODO/ADR for this.
- The resolver fallback (`pkg/tenant/resolver.go`) should append `sslmode=require`
  only if no sslmode is present in the URL. The plan says this correctly.

**Recommendation:** Accept `sslmode=require` for now with documented constraint.
Add to `deploy/secrets/README.md`: "If PG moves off localhost, upgrade to
verify-full."

---

### C5 -- NATS Per-Service Auth

**Proposed:** Complete `nats-server.conf` with users for agent, traces, extractor,
platform.

**Assessment: CORRECT direction, missing permissions.**

Current config in `deploy/nats/nats-server.conf` has 6 users. Missing:
- `agent` -- needs: publish `tenant.*.traces.*`, `tenant.*.notify.*`; subscribe none
- `traces` -- needs: subscribe `tenant.*.traces.>`; publish none (or `tenant.*.notify.*` for alerts)
- `extractor` -- needs: subscribe `tenant.*.ingest.>`; publish `tenant.*.ingest.*`
- `platform` -- needs: publish `tenant.*.notify.*`, `tenant.*.module.*`; subscribe none

**Issues with existing config:**
1. **WS Hub subscribes to `tenant.>`** (line 43-47). This is correct -- the Hub
   bridges to browsers. But it means a compromised WS Hub can read ALL tenant
   events. Acceptable given its role, but document this trust boundary.
2. **No deny rules.** NATS authorization with `permissions` uses allowlists, which
   is correct. But verify that services without subscribe permissions cannot
   subscribe to anything. In NATS, if `subscribe` is an empty list `[]`, no
   subscriptions are allowed. This is correct.
3. **Feedback service** is listed but feedback service doesn't exist in
   `services/*/cmd/main.go` -- it exists. Verified.

**Additional gap:** The plan says "Montar `nats-server.conf` en el container NATS"
but doesn't mention that `$AUTH_NATS_PASS` etc. use NATS variable substitution.
Verify that NATS server supports env var interpolation in config. NATS does
support this natively, so this is fine.

---

### H8 -- GetSession Ownership Check

**Proposed:** Add `AND user_id = $2` to `GetSession` query in chat.sql.

**Assessment: CORRECT but the service layer already compensates.**

Looking at the code:
- `chat.sql:8-9`: `GetSession` filters only by `id` -- no user_id
- `chat/internal/service/chat.go:93-106`: GetSession fetches by ID, then checks
  `row.UserID != userID` and returns `ErrNotOwner`

The service-layer check is a valid pattern (fetch-then-validate) and it works.
The plan's suggestion to push this into the query is also valid (defense in depth,
prevents future regressions). **Both approaches are secure today.** The query-level
fix is better long-term.

**Other queries with the same pattern (plan missed):**

1. **`TouchSession`** (`chat.sql:22-23`): `UPDATE sessions SET updated_at = now() WHERE id = $1`
   -- No user_id filter. Called from `AddMessage` AFTER ownership is verified, so
   currently safe. But if TouchSession is ever called directly, it's vulnerable.
   Add `AND user_id = $2` for defense-in-depth.

2. **`GetJob` in ingest** (`ingest.sql:6-8`): `SELECT ... FROM ingest_jobs WHERE id = $1`
   -- No user_id. The service layer checks ownership
   (`ingest/internal/service/ingest.go:197-210`). Same pattern as chat. Fix at query
   level.

3. **`GetDocument`** (`documents.sql:5`): `SELECT * FROM documents WHERE id = $1`
   -- No uploaded_by filter. No ownership check found in the service layer. This
   means ANY authenticated user in the same tenant can read any document by ID.
   **This is a design decision** (documents are shared within tenant), but it should
   be documented explicitly.

4. **`DeleteDocument`** (`documents.sql:31`): `DELETE FROM documents WHERE id = $1`
   -- No ownership check. Any user with the right permissions can delete any document
   in the tenant. Same design-decision question.

5. **`ListDocuments`** (`documents.sql:16`): No user filter at all. Returns all
   documents in the tenant DB. This is correct for a shared-tenant design but
   **has no pagination bounds** -- the limit is a param, but there's no max cap.

**Recommendation:** Fix GetSession, TouchSession, GetJob at query level. Document
the document-sharing model as intentional.

---

### H11 -- JWT JTI Auto-generation

**Proposed:** If `claims.ID` is empty in `CreateAccess`/`CreateRefresh`, generate
with `uuid.NewString()`.

**Assessment: CORRECT but the code already does this in the service layer.**

Looking at `auth/internal/service/auth.go`:
- Line 214: `accessClaims.ID = uuid.New().String()`
- Line 217: `refreshClaims.ID = uuid.New().String()`
- Line 182: MFA token: `mfaClaims.ID = uuid.New().String()`
- Line 321: Post-MFA: `accessClaims.ID = uuid.New().String()`
- Line 323: `refreshClaims.ID = uuid.New().String()`
- Line 417: Refresh: `newAccessClaims.ID = uuid.New().String()`
- Line 420: `newRefreshClaims.ID = uuid.New().String()`

Every token creation path already sets JTI. The H11 fix is still valuable as
defense-in-depth (prevents regressions), but it's not a vulnerability today.

**Edge case the plan misses:** The blacklist check in `pkg/middleware/auth.go:69`
only fires when `claims.ID != ""`. If a token somehow has no JTI (e.g., external
tooling), it bypasses the blacklist silently. The H11 fix resolves this, but the
middleware should ALSO reject tokens without JTI:

```go
if cfg.Blacklist != nil && claims.ID == "" {
    writeJSONError(w, http.StatusUnauthorized, "invalid token: missing jti")
    return
}
```

**Severity of edge case: MEDIUM** (requires forging a token without JTI, which
requires the private key, which means the attacker already has everything).

---

### M14 -- API Key Cache

**Proposed:** Exclude `APIKey` from Redis cache, or use a separate struct, or
encrypt with `pkg/crypto`.

**Assessment: CORRECT concern, wrong default recommendation.**

Looking at `pkg/config/resolver.go:176-179`:
```go
data, _ := json.Marshal(mc) // mc includes APIKey field
r.cache.Set(ctx, "sda:model:"+modelID, string(data), r.cacheTTL)
```

The `ModelConfig` struct has `APIKey string json:"api_key,omitempty"` and the
full struct (including APIKey) is cached in Redis as JSON.

**Options ranked:**
1. **Best: Separate struct** -- `CachedModelConfig` without APIKey. Read APIKey
   from DB only when needed (per-request to LLM). Cache hit rate stays the same
   for endpoint/model_id/costs, APIKey is never in Redis.
2. **Acceptable: Encrypt** -- Use `pkg/crypto.Encrypt()` before caching. Adds
   complexity (key management for crypto), decryption on every cache hit.
3. **Worst: Exclude** -- Don't cache at all. Every LLM call hits the DB.

**Recommendation:** Option 1 (separate struct). Clean, simple, no crypto overhead.

---

## 2. Missed Vulnerabilities (Not in Plan 08)

### MISSED-1: CRITICAL -- Search Service Has No Tenant Isolation in SQL

**File:** `services/search/internal/service/search.go:116-134`

The `loadTrees` function queries `document_trees` and `documents` but does NOT
filter by tenant. The search service connects to a single tenant DB
(`POSTGRES_TENANT_URL`), which provides implicit isolation in dev. But in
multi-tenant mode (if search ever gets a Resolver), this would leak cross-tenant.

More importantly: the search service uses `pool.Query(ctx, query, args...)` with
raw SQL strings, not sqlc. While the SQL is parameterized (no injection risk),
the raw queries bypass sqlc's compile-time verification.

**Current risk:** LOW in dev (single tenant DB). CRITICAL if search migrates to
multi-tenant resolver.

**Fix:** Either:
a. Document that search MUST always connect to a per-tenant DB (never platform DB)
b. Add tenant_id filter to all queries as defense-in-depth

---

### MISSED-2: HIGH -- Agent Service Has WriteTimeout: 0 (Plan Mentions, But Understates)

**File:** `services/agent/cmd/main.go:164`

`WriteTimeout: 0` means the server will wait indefinitely for a response to be
written. This enables slowloris attacks where a client opens a connection and
reads the response body one byte at a time, forever, tying up a goroutine.

The plan lists this as M11 (medium) and proposes 5 minutes. This should be
HIGH because it's a denial-of-service vector. The proposed fix (5 min) is correct
for streaming SSE, but the plan should also mandate `ReadHeaderTimeout` (currently
absent from ALL services):

```go
ReadHeaderTimeout: 10 * time.Second,
```

Without `ReadHeaderTimeout`, a client can open a TCP connection and send headers
one byte at a time indefinitely. This is separate from `ReadTimeout`.

---

### MISSED-3: HIGH -- Blacklist Fail-Open Policy

**File:** `pkg/middleware/auth.go:71-74`

```go
if err != nil {
    slog.Error("blacklist check failed", "error", err)
    // fail-open: if Redis is down, don't block all requests
}
```

If Redis goes down, ALL revoked tokens become valid again. The comment says
this is intentional, but for a security-critical feature (logout), fail-open
is dangerous.

**Recommendation:** Make the fail policy configurable:
- `FailOpen: true` (current, for availability)
- `FailOpen: false` (for security, rejects all requests when Redis is down)

For auth service specifically, fail-closed is correct (a few minutes of
downtime is better than accepting revoked tokens). For non-auth services,
fail-open is acceptable.

---

### MISSED-4: MEDIUM -- Platform Service CreateTenant Accepts Raw Connection URLs

**File:** `services/platform/internal/handler/platform.go:122-156`

The `CreateTenant` endpoint accepts `postgres_url` and `redis_url` in the
request body and stores them directly. A platform admin could set these to
point to arbitrary hosts, enabling:
- SSRF: pointing postgres_url to an internal service
- Data exfil: pointing to an attacker-controlled PostgreSQL
- Credential harvest: if the URL is logged anywhere

**Fix:**
1. Validate URL format (must be postgres:// or postgresql://)
2. Reject private IP ranges unless explicitly allowed
3. Or better: don't accept URLs via API at all -- generate them from a template
   (e.g., `postgres://sda:{generated_pass}@postgres:5432/sda_tenant_{slug}`)

---

### MISSED-5: MEDIUM -- Notification Service Limit Parameter Unbound

**File:** `services/notification/internal/handler/notification.go:61`

```go
limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
```

If `limit` fails to parse, it's 0. Then it goes to the service layer unchecked.
Looking at the query: `ListNotifications` uses `LIMIT $2`. With limit=0, this
returns 0 rows (harmless). But with limit=999999999, it returns everything.

**Fix:** Cap limit to a reasonable max (e.g., 100) and set a default (e.g., 50).
This is the same pagination gap that H4 addresses, but the plan doesn't
explicitly mention notification's limit parameter.

---

### MISSED-6: MEDIUM -- auth.go:122-125 JSON Injection in writeJSONError

**File:** `pkg/middleware/auth.go:122-126`

```go
w.Write([]byte(`{"error":"` + msg + `"}`))
```

The plan mentions this as L4 (low). But if `msg` contains a double quote or
backslash, the JSON becomes malformed or injectable. While the callers currently
pass hardcoded strings ("missing authorization", "invalid token", "token revoked",
"mfa verification required", "tenant mismatch"), any future caller passing
user-controlled data would create a JSON injection.

**This should be MEDIUM, not LOW.** The fix is trivial (`json.Marshal`), but the
blast radius of getting it wrong in the future is high.

---

## 3. Priority/Severity Assessment

### Items that should be upgraded:

| Finding | Plan severity | Correct severity | Reason |
|---------|--------------|-----------------|--------|
| M11 (WriteTimeout: 0) | Medium | HIGH | DoS vector, single request can tie up goroutine forever |
| L4 (writeJSONError) | Low | MEDIUM | JSON injection risk with any future non-hardcoded msg |
| L5 (DeleteJobByID) | Low | MEDIUM | Authorization bypass -- any service-internal caller can delete jobs without ownership |
| M9 (gosec || true) | Medium | HIGH | Security scanner results are silently ignored; defeats purpose of having scanner |

### Items correctly prioritized:

- C1 (audit logging): Correct critical. Compliance blocker.
- C2 (rate limiting): Correct critical. But see gaps above.
- C3 (sqlc paths): Correct critical for build reliability.
- C4 (sslmode): Correct critical for prod.
- C5 (NATS auth): Correct critical. Single shared token is unacceptable.
- H8 (session ownership): Correct high.
- H11 (JWT JTI): Correct high, though already mitigated.
- H4 (pagination): Correct high for scalability+security (unbounded queries).

---

## 4. gRPC Phase (Fase 4) -- Security Surface Analysis

The plan's gRPC section (Fase 4) has **zero security design**. This is a
significant gap for a hardening plan.

### New attack surfaces introduced by gRPC:

**4.1 Authentication between services:**
- Currently, inter-service HTTP calls (e.g., Agent -> Search) use no auth or
  forward the user's JWT. With gRPC, the same pattern applies but must be
  explicit.
- **Question:** Does the gRPC call forward the user's JWT as metadata?
  Or does it use a service-to-service token?

**Recommendation:** Forward user JWT as gRPC metadata for user-context calls.
For system-level calls (Platform -> Auth tenant lifecycle), use a service
account JWT signed with the same Ed25519 key.

**4.2 mTLS between services:**
- In the current architecture (all on one host), mTLS is optional. Docker network
  provides basic isolation.
- **If services ever split across hosts:** mTLS becomes mandatory.
- **Recommendation for now:** No mTLS needed (same host). Add a TODO/ADR.

**4.3 gRPC server ports exposed:**
- The plan says "each service exposes a gRPC server on a separate port (e.g.,
  auth HTTP=8001, gRPC=9001)." These ports must NOT be exposed to Traefik or
  the internet.
- **Fix:** Bind gRPC servers to the Docker internal network only. No port
  mapping in docker-compose. Only HTTP ports are exposed via Traefik.

**4.4 Input validation on gRPC:**
- Protobuf provides type safety but not business validation.
- gRPC handlers must validate inputs the same way HTTP handlers do.
- **Fix:** Add interceptor (middleware) for:
  - Request size limits (equivalent to MaxBytesReader)
  - Auth token verification (equivalent to Auth middleware)
  - Rate limiting

**4.5 Reflection and server info:**
- gRPC reflection should be disabled in prod (it reveals all RPCs).
- Health checking should use the standard gRPC health protocol.

**Recommendation:** Add a "Paso 0: Security Design" to Fase 4 that covers
auth forwarding, port binding, interceptors, and reflection.

---

## 5. Backup Strategy (Fase 8) -- Security Assessment

### L15 -- Backup

**Proposed:** pg_dump per tenant, gzip, upload to MinIO, 30-day/12-month retention.

**Assessment: INCOMPLETE on security.**

**5.1 Encryption at rest:**
The plan does NOT mention encrypting backups. A pg_dump contains:
- All user data (emails, names)
- Password hashes (bcrypt, but still sensitive)
- MFA secrets (if not encrypted at app layer)
- API keys
- Tenant connection URLs (which contain credentials)

**Fix:** Add `gpg --encrypt` or `age` encryption BEFORE uploading to MinIO.
The encryption key should be stored separately from MinIO (e.g., in a password
manager or hardware security module for prod).

```bash
pg_dump ... | gzip | gpg --encrypt --recipient backup@sda.app > backup.sql.gz.gpg
```

**5.2 Access control:**
- MinIO bucket `sda-backups/` must have restricted access policies.
- The backup script needs credentials to pg_dump (superuser) and MinIO (write).
- These credentials should be stored in Docker secrets, not env vars.
- MinIO should have a separate user for backups with write-only access to the
  bucket (no delete). This prevents a compromised backup script from deleting
  existing backups (ransomware protection).

**5.3 Restore script security:**
- The restore script (`deploy/scripts/restore.sh`) will need superuser
  credentials. Document that this script should NOT be deployed to prod
  containers -- it should run manually from a trusted workstation.

**5.4 Backup integrity:**
- Add SHA-256 checksums alongside each backup file.
- The restore script should verify the checksum before restoring.

**5.5 Redis backup:**
- Redis contains the token blacklist and config cache. The blacklist has
  TTL-based entries that are inherently ephemeral. Backing up Redis `.rdb`
  is fine but not critical -- worst case after restore, some recently-revoked
  tokens are valid until they expire naturally (15 min max).

**5.6 NATS JetStream:**
- Including `/data` is correct. JetStream messages in flight should survive
  a backup/restore cycle.

**Severity of gaps: HIGH (unencrypted backups with credentials).**

---

## 6. Tenant Isolation Audit

### Current state:

| Layer | Isolation mechanism | Status |
|-------|-------------------|--------|
| JWT | Claims contain tid + slug | GOOD -- validated in Verify() |
| Subdomain | Traefik extracts slug, middleware cross-validates | GOOD |
| Header stripping | auth.go strips X-* headers before processing | GOOD |
| Database | Per-tenant PostgreSQL via Resolver | GOOD -- $1 parameterized |
| Redis | Per-tenant via Resolver.RedisClient() | GOOD |
| NATS subjects | Validated via IsValidSubjectToken() | GOOD |
| SQL queries | All use parameterized queries (sqlc + pgx) | GOOD |
| Search service | Connects to single tenant DB | ACCEPTABLE for dev |
| Traces service | tenant_id from JWT, validated in queries | GOOD |
| Traefik prod | strip-spoofed-headers + tenant-from-subdomain | GOOD |

### Gaps:

1. **Search service loadTrees():** Uses raw SQL without tenant filter. Currently
   safe because it connects to per-tenant DB. Document this constraint.

2. **Platform service:** Operates on platform DB (all tenants). The
   `requirePlatformAdmin` middleware checks `claims.Slug == h.platformSlug`,
   which is correct. But the platform slug is passed as a constructor arg.
   If misconfigured, any admin from any tenant could access platform endpoints.

3. **Blacklist is not per-tenant:** `pkg/security/blacklist.go` uses key
   `sda:token:blacklist:{jti}`. If all tenants share the same Redis (current
   dev setup), a token JTI from tenant A could collide with tenant B (UUID
   collision is astronomically unlikely, but the key schema should include
   tenant for correctness). The Revoke method signature accepts `jti` only.

**Overall tenant isolation verdict: GOOD for current architecture. No
cross-tenant vectors found.**

---

## 7. Missing Security Items (Spec vs Reality)

| Spec promise | Status | Plan 08 addresses? |
|-------------|--------|-------------------|
| Audit logging | Code exists, partially wired in auth+chat+search | YES (C1) |
| Rate limiting | Not implemented in Go; exists in Traefik prod | YES (C2) |
| JWT blacklist | Implemented + wired in auth | YES (H11 improvement) |
| NATS per-service auth | Config exists, not deployed | YES (C5) |
| SSL on PG | Not enabled | YES (C4) |
| CrowdSec IDS | Not deployed | YES (M12 in Fase 8) |
| Docker socket proxy | Not deployed | YES (L14 in Fase 8) |
| CPU limits | Not set | YES (L13 in Fase 8) |
| CORS | Configured in Traefik prod | Already done |
| WS_ALLOWED_ORIGINS in prod | Not verified | Plan doesn't mention |
| Distroless containers | Not verified | Plan doesn't mention |
| Non-root containers | Not verified | Plan doesn't mention |
| Network segmentation per tenant | Not implemented | Plan doesn't mention |
| Secrets in Docker secrets (not env) | Dev uses env vars | Plan mentions L2 (comment only) |
| HTTPS in prod | Traefik has TLS config | Already done |

---

## 8. Dependency CVEs

| Package | Version | Known CVEs |
|---------|---------|-----------|
| golang-jwt/jwt/v5 | v5.3.1 | None known (latest) |
| go-chi/chi/v5 | v5.2.5 | None known (latest) |
| jackc/pgx/v5 | v5.9.1 | None known (latest) |
| nats-io/nats.go | v1.50.0 | None known (latest) |
| redis/go-redis/v9 | v9.18.0 | None known (latest) |
| coder/websocket | (used in ws) | Verify version -- no CVEs in latest |

All major dependencies are at latest versions. No known CVEs as of 2026-04-05.

---

## 9. Additional Findings Per Phase

### Fase 1 corrections needed:
- C2: Add refresh and MFA rate limiting
- C4: Document sslmode decision explicitly

### Fase 2 corrections needed:
- H11: Also reject tokens without JTI in middleware
- L4: Upgrade to MEDIUM severity
- L5: Upgrade to MEDIUM severity
- M9: Upgrade to HIGH severity
- Add: TouchSession ownership check (chat.sql)

### Fase 4 mandatory additions:
- Security design document (auth, port binding, interceptors)
- No gRPC reflection in prod
- gRPC ports bound to internal network only

### Fase 8 mandatory additions:
- Backup encryption (gpg/age)
- Backup access control (write-only MinIO user)
- Backup integrity verification (SHA-256)
- Restore script security (manual only, not deployed)

---

## 10. Summary of All Findings

### CRITICAL (3 -- plan's 5 are correct + 0 new)
- C1. Audit logging unwired -- **plan correct**
- C2. Rate limiting missing -- **plan correct but incomplete (add refresh, MFA)**
- C3. sqlc paths broken -- **plan correct**
- C4. sslmode=disable in prod -- **plan correct but document constraint**
- C5. NATS single shared token -- **plan correct, add missing service perms**

### HIGH (8 -- plan's 11 reclassified + 3 new)
- H8. GetSession no ownership in query -- **plan correct**
- H11. JTI auto-gen -- **plan correct, add middleware reject on empty JTI**
- M9. gosec/trivy silently pass -- **upgrade from medium to HIGH**
- M11. Agent WriteTimeout: 0 -- **upgrade from medium to HIGH**
- MISSED-2. All services missing ReadHeaderTimeout -- **NEW HIGH**
- MISSED-3. Blacklist fail-open -- **NEW HIGH**
- MISSED-7. Backup unencrypted -- **NEW HIGH**
- H4, H9, H2, H7, H10, H3, H5, H6 -- **plan correct as-is**

### MEDIUM (8 -- plan's 15 + 3 new, minus upgrades)
- M14. API key in Redis cache -- **plan correct, recommend separate struct**
- L4. writeJSONError string concat -- **upgrade from low to MEDIUM**
- L5. DeleteJobByID no ownership -- **upgrade from low to MEDIUM**
- MISSED-1. Search no tenant filter in SQL -- **NEW MEDIUM**
- MISSED-4. CreateTenant accepts raw URLs -- **NEW MEDIUM**
- MISSED-5. Notification limit unbound -- **NEW MEDIUM**
- MISSED-6. writeJSONError (same as L4 above)
- M1, M3, M4, M6, M7, M8, M10, M15 -- **plan correct as-is**

### LOW (plan's 21 minus upgrades)
- All remaining lows are correctly prioritized.

---

## Veredicto

**NO APTO para produccion** en su forma actual.

El plan es un excelente punto de partida pero necesita las correcciones
documentadas en esta auditoria antes de ejecutar cada fase. Los bloqueantes
principales son:

1. Rate limiting para refresh y MFA (brute force vectors)
2. gRPC Fase 4 sin diseno de seguridad
3. Backups sin encripcion
4. ReadHeaderTimeout ausente en todos los servicios
5. Fail-open en blacklist deberia ser configurable

Una vez incorporadas estas correcciones al plan, las 8 fases pueden ejecutarse
en orden y el sistema estaria listo para usuarios reales.
