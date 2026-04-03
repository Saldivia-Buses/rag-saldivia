# Security Audit -- PR #53 Feedback Service + OTel Instrumentation

**Date:** 2026-04-03
**Branch:** feat/wiring-polish-feedback
**Auditor:** security-auditor agent (Opus)
**Scope:** Feedback service (NATS consumer, aggregator, alerter, HTTP handlers), OTel pkg, OTel instrumentation across all services

---

## Executive Summary

The Feedback Service is well-structured with parameterized SQL, proper JWT middleware wiring, and header-stripping anti-spoofing. However, there are **2 critical** and **3 high** severity findings that must be addressed before merge. The most dangerous is NATS running without authentication (any process on the network can publish to any tenant's feedback subjects), and the LIMIT clause concatenation in the Quality handler.

---

## CRITICAL (block merge)

### C-1. NATS has zero authentication -- any producer can inject events for any tenant

**Files:**
- `deploy/docker-compose.dev.yml:68-81` -- NATS started with `--js --sd /data -m 8222`, no auth flags
- No `nats.conf` or `nats-server.conf` anywhere in `deploy/`
- `services/feedback/internal/service/consumer.go:17` -- subscribes to `tenant.*.feedback.>`

**Impact:** NATS has no auth, no accounts, no subject-level ACL. Any process that can reach port 4222 (which is bound to `0.0.0.0:4222` in docker-compose) can publish to `tenant.saldivia.feedback.error_report` and inject fake feedback events into any tenant's database. In production, this is remote code execution equivalent for data integrity -- an attacker can fabricate quality scores, error spikes, and trigger alerts for any tenant.

The consumer extracts tenant from the NATS subject (line 118: `parseSubject(msg.Subject())`), not from any authenticated identity. There is no verification that the publisher is authorized to publish to that tenant's subjects.

**Fix (before merge -- at least the port binding):**
1. Stop binding NATS port `4222` to `0.0.0.0` in dev compose. Use `127.0.0.1:4222:4222` or remove the port binding entirely (services in Docker network can still reach it).
2. Create a production `nats.conf` with accounts and per-service publish/subscribe permissions. Each service should only be allowed to publish to subjects it owns (e.g., chat service can only publish to `tenant.*.feedback.response_quality`).
3. Add a TODO/issue for NATS auth before production. This is the single most dangerous finding.

**Severity justification:** In the current architecture, the feedback consumer trusts NATS subject routing as the tenant isolation boundary. Without NATS auth, there IS no tenant isolation for feedback data.

---

### C-2. SQL LIMIT clause concatenated from user input (uncapped)

**File:** `services/feedback/internal/handler/feedback.go:137`

```go
query += ` ORDER BY created_at DESC LIMIT ` + strconv.Itoa(limit)
```

The `limit` value comes from `parseIntParam(r.URL.Query().Get("limit"), 50)` (line 124). While `parseIntParam` (line 281-290) does reject non-positive values and non-integers (safe from SQL injection since `strconv.Itoa` always produces a clean integer), it has **no upper bound**. A request with `?limit=999999999` will attempt to load ~1 billion rows into memory.

This is both a DoS vector and a departure from the parameterized SQL pattern used everywhere else in this service.

**Fix:**
```go
func parseIntParam(s string, fallback, max int) int {
    if s == "" {
        return fallback
    }
    v, err := strconv.Atoi(s)
    if err != nil || v <= 0 {
        return fallback
    }
    if v > max {
        return max
    }
    return v
}
```

Call with `parseIntParam(r.URL.Query().Get("limit"), 50, 200)`.

Also, use a parameterized `$N` placeholder for LIMIT instead of concatenation:
```go
query += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(len(args)+1)
args = append(args, limit)
```

This applies to `feedback.go:137` (Quality handler). The Errors handler at line 186 correctly uses `LIMIT $2` as a parameter -- follow that pattern.

---

## HIGH (fix before production)

### H-1. Alerter called with empty tenant slug -- Notify will fail or publish to malformed subject

**File:** `services/feedback/internal/service/aggregator.go:112`

```go
a.alerter.CheckAndAlert(ctx, tenantID, "", scores) // slug resolved from tenantID in prod
```

The empty string `""` is passed as `tenantSlug`. In `alerter.go:194`, this gets passed to `a.publisher.Notify(tenantSlug, ...)`. In `pkg/nats/publisher.go:39`, `isValidSubjectToken("")` returns `false`, so the publish will fail silently (error returned but the alerter at line 194 does not check the return value of `Notify`).

**Impact:** All feedback alerts in production are silently dropped. No notifications will ever be sent for health threshold violations. The comment says "slug resolved from tenantID in prod" but there is no code that does this.

**Fix:**
1. The aggregator must resolve the tenant slug from the platform DB before calling `CheckAndAlert`. Query `SELECT slug FROM tenants WHERE id = $1`.
2. The alerter must check and log the error from `publisher.Notify()`.

---

### H-2. Platform admin check relies solely on X-User-Role header string comparison, not role hierarchy

**File:** `services/feedback/internal/handler/platform_feedback.go:33-35,88-91,160-162`

```go
role := r.Header.Get("X-User-Role")
if role != "admin" {
    writeJSON(w, http.StatusForbidden, ...)
}
```

**Analysis:** This header IS safe from spoofing because:
- The `sdamw.Auth(jwtSecret)` middleware is applied in `cmd/main.go:149` before the platform routes mount.
- `pkg/middleware/auth.go:27-31` strips all identity headers (`X-User-Role`, etc.) before processing.
- `pkg/middleware/auth.go:55` re-sets `X-User-Role` from the JWT claims.

So the header comes from a verified JWT. However, the check is a simple string comparison `!= "admin"`. If additional admin-tier roles are added later (e.g., `super_admin`, `platform_admin`), they will be incorrectly rejected. This is not a vulnerability today, but it is a fragile pattern.

**Fix:** Use a role-checking helper that understands the role hierarchy:
```go
func isAdminRole(role string) bool {
    return role == "admin" || role == "super_admin"
}
```

Or better: define an `AllowedRoles` middleware in `pkg/middleware/` that works like `requirePlatformAdmin` in the platform service.

**Severity note:** Downgraded from critical to high because the header IS sourced from JWT via middleware. The risk is maintainability, not current exploitability.

---

### H-3. NATS payload stored as-is in JSONB context column -- no sanitization of sensitive data

**File:** `services/feedback/internal/service/consumer.go:147`

```go
Context: msg.Data(), // store the full original payload as context
```

The entire NATS message payload is persisted as JSONB in the `feedback_events.context` column. If any producing service includes sensitive data in the payload (e.g., a user's session token in an error report, a password in a security event, an API key in a performance trace), it gets stored permanently in the tenant database and is queryable via the REST API (`GET /v1/feedback/errors` returns the `context` field at line 201).

**Impact:** Data leak through persistence. The Errors handler at line 184-201 returns the raw `context` JSONB to any authenticated user of the tenant.

**Fix:**
1. Define a sanitization function that strips known-sensitive keys from the payload before storage:
```go
var sensitiveKeys = []string{"password", "token", "secret", "api_key", "authorization", "cookie"}

func sanitizeContext(data []byte) json.RawMessage {
    var m map[string]any
    if err := json.Unmarshal(data, &m); err != nil {
        return json.RawMessage("{}")
    }
    for _, key := range sensitiveKeys {
        delete(m, key)
        delete(m, strings.ToUpper(key))
    }
    out, _ := json.Marshal(m)
    return out
}
```
2. Apply before storage: `Context: sanitizeContext(msg.Data())`

---

## MEDIUM (backlog priority)

### M-1. No MaxBytesReader on HTTP handlers

**Files:** All handlers in `services/feedback/internal/handler/`

None of the handlers use `http.MaxBytesReader`. While all current endpoints are GET-only (no request body parsing), this should be added defensively. If a POST endpoint is added later without it, a large payload could exhaust memory.

**Fix:** Add MaxBytesReader as middleware in `cmd/main.go` or as the first line of any future POST handler.

---

### M-2. No upper bound on parseIntParam -- DoS via large limit

Already covered in C-2 for the LIMIT concatenation. Even in handlers that use parameterized LIMIT (like `Errors` at line 186 which uses `$2`), there is no cap. `?limit=10000000` will still attempt to load 10M rows.

**Fix:** Add `max` parameter to `parseIntParam` (see C-2 fix). Apply to all callsites:
- `feedback.go:124` (Quality)
- `feedback.go:180` (Errors)
- `platform_feedback.go:98` (Alerts)

---

### M-3. Feedback service not routed through Traefik

**Files:**
- `deploy/traefik/dynamic/dev.yml` -- no `feedback` router or service entry
- `deploy/traefik/dynamic/prod.yml` -- no `feedback` router or service entry
- `deploy/docker-compose.dev.yml` -- no feedback container definition

The feedback service runs on `:8008` but there is no Traefik routing rule for `/v1/feedback` or `/v1/platform/feedback`. This means:

1. In dev: the service is accessible only by direct port access, bypassing Traefik entirely. This is fine for dev but means no rate limiting or CORS.
2. In prod: the service is completely unreachable from the frontend unless Traefik config is updated.

**Fix:** Add routing rules to both Traefik configs:
```yaml
# dev.yml
feedback:
  rule: "PathPrefix(`/v1/feedback`)"
  service: feedback
  entryPoints: [web]
  middlewares: [dev-tenant]

# prod.yml
feedback:
  rule: "PathPrefix(`/v1/feedback`)"
  service: feedback
  entryPoints: [web]
  middlewares: [strip-spoofed-headers, tenant-from-subdomain, rate-limit, cors]

platform-feedback:
  rule: "Host(`platform.sda.app`) && PathPrefix(`/v1/platform/feedback`)"
  service: feedback
  entryPoints: [web]
  middlewares: [rate-limit, cors]
```

Note: The platform feedback routes (`/v1/platform/feedback/*`) currently might collide with the platform service's `/v1/platform/*` Traefik route. This needs a priority rule or the feedback platform endpoints should be served through the platform service Traefik entry.

---

### M-4. Aggregator uses env var TENANT_ID -- single-tenant limitation

**File:** `services/feedback/cmd/main.go:119`

```go
tenantID := env("TENANT_ID", "dev")
```

The aggregator hardcodes a single tenant ID from an environment variable. In a multi-tenant deployment, you need one feedback service instance per tenant, or the aggregator must iterate over all tenants. Currently, only one tenant's health score gets computed.

**Fix:** The aggregator should query the platform DB for active tenants and loop over each. This is a design issue, not a security vulnerability, but it means health scores and alerts only work for one tenant.

---

### M-5. OTel OTLP exporter uses WithInsecure (no TLS to collector)

**File:** `pkg/otel/otel.go:50`

```go
otlptracegrpc.WithInsecure(), // TLS not needed for local collector
```

This is acceptable when the OTel Collector runs on the same host or Docker network. In production, if the collector is on a separate machine, trace data (which includes HTTP paths, status codes, and timing) would travel unencrypted.

**Fix:** Make TLS configurable:
```go
if cfg.Insecure {
    opts = append(opts, otlptracegrpc.WithInsecure())
} else {
    opts = append(opts, otlptracegrpc.WithTLSCredentials(/* ... */))
}
```

---

## LOW (nice to have)

### L-1. OTel otelhttp captures HTTP route, method, status but not sensitive data

**File:** `services/feedback/cmd/main.go:156`

```go
Handler: otelhttp.NewHandler(r, "sda-feedback"),
```

**Verification result:** The `otelhttp` handler by default captures: HTTP method, route, status code, content length, and user agent. It does NOT capture request/response bodies, Authorization headers, or cookies. This is safe. No sensitive data is leaked into traces.

However, the `otelhttp` instrumentation is applied to all 8 services simultaneously in this PR. This is a wide blast radius. If any service has custom span attributes added later that include sensitive data, it would be captured.

**Recommendation:** Document in the OTel README that span attributes must NEVER include tokens, passwords, PII, or request bodies.

---

### L-2. Consumer Fetch batch size is hardcoded

**File:** `services/feedback/internal/service/consumer.go:101`

```go
batch, err := c.cons.Fetch(10, jetstream.FetchMaxWait(5e9))
```

Batch size 10 with 5s wait is reasonable. MaxDeliver is 5 (line 69). This provides natural backpressure -- if the consumer can't keep up, messages stay in JetStream. The 7-day MaxAge (line 59) prevents unbounded growth.

This is adequate for now but should be tunable via env var for production load.

---

### L-3. Error rows.Scan not checked in several handlers

**Files:**
- `feedback.go:154` -- `rows.Scan(...)` with no error check
- `feedback.go:241` -- same
- `platform_feedback.go:64` -- same
- `platform_feedback.go:123-124` -- same
- `platform_feedback.go:199` -- same

Scan errors are silently ignored. While this won't cause security issues (scan errors produce zero-value fields), it makes debugging hard and could mask data corruption.

**Fix:** Check all `rows.Scan` return values.

---

### L-4. QueryRow.Scan errors not checked in summary handler

**File:** `feedback.go:52-59,63-69,72-78,82-89`

Multiple `QueryRow().Scan()` calls where the error is silently dropped. If the query fails, zero values are returned to the user, which could be misleading.

---

## Tenant Isolation Audit

### Result: ACCEPTABLE with caveats

**HTTP layer:** SAFE. The feedback service uses `sdamw.Auth(jwtSecret)` middleware (main.go:142,149) which:
1. Strips spoofed headers (auth.go:27-31)
2. Verifies JWT
3. Sets identity headers from JWT claims

Tenant-scoped handlers query from `tenantDB` which is a per-tenant PostgreSQL database. There is no `WHERE tenant_id = ...` needed because the database itself IS the isolation boundary. This is correct.

**Platform handlers:** SAFE. Platform handlers query from `platformDB` and enforce `X-User-Role == "admin"` which comes from a verified JWT.

**NATS layer:** NOT SAFE. See C-1. NATS has no auth, so any process can publish to any tenant's feedback subjects. The consumer trusts the subject routing as tenant identification (consumer.go:118) without any cross-verification.

**Aggregator cross-DB flow:** SAFE for data leakage. The aggregator reads counts and averages from tenant DB (e.g., `COUNT(*)`, `AVG(score)`, percentiles) and writes them to platform DB as numeric scores. It does NOT copy event text, comments, user data, or context fields to the platform DB. The platform tables (`feedback_metrics`, `tenant_health_scores`) contain only: tenant_id, module, category, period, numeric counts, and numeric scores. This is aggregated telemetry, not sensitive data.

---

## Missing Security Controls

| Control | Status | Notes |
|---|---|---|
| NATS authentication | MISSING | C-1 above |
| NATS subject-level ACL | MISSING | No per-service publish restrictions |
| Rate limiting on feedback endpoints | MISSING | Not routed through Traefik (M-3) |
| Payload sanitization in consumer | MISSING | H-3 above |
| Audit log for platform admin access | MISSING | Platform feedback endpoints have no audit trail |
| Feedback service in Traefik config | MISSING | M-3 above |
| Consumer payload size limit | MISSING | JetStream has default 1MB max per message, acceptable |
| Input validation on query params | PARTIAL | Period is allowlisted, limit is unbounded (M-2) |

---

## CVE Check

| Dependency | Version | Known CVEs |
|---|---|---|
| `nats.go` | v1.50.0 | None known (latest as of 2026-04) |
| `pgx/v5` | v5.9.1 | None known |
| `chi/v5` | v5.2.5 | None known |
| `otel` | v1.42.0 | None known |
| `otelhttp` | v0.67.0 | None known |
| `golang-jwt/jwt/v5` | (in pkg/jwt) | None known |

All dependencies are at recent versions. No known CVEs.

---

## Veredicto: NO APTO para merge sin fixes

### Must fix before merge (2 critical):
1. **C-1:** Bind NATS port to localhost in dev compose; create issue for NATS auth in prod
2. **C-2:** Cap LIMIT parameter and use parameterized placeholder

### Must fix before production (3 high):
3. **H-1:** Resolve tenant slug for alerter notifications
4. **H-2:** Use role hierarchy helper for admin check
5. **H-3:** Sanitize NATS payload before persisting to context column

### Should fix before production (5 medium):
6. **M-1:** MaxBytesReader on handlers
7. **M-2:** Cap all limit parameters
8. **M-3:** Add Traefik routing for feedback service
9. **M-4:** Multi-tenant aggregator support
10. **M-5:** Configurable TLS for OTel exporter
