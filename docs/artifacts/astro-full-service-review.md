# Astro Service — Full Integral Review

**Fecha:** 2026-04-06
**Resultado:** CAMBIOS REQUERIDOS
**Reviewer:** gateway-reviewer (Opus 4.6)
**Scope:** Complete service review — 19 PRs, 40 source files, ~2800 LoC

---

## Executive Summary

The astro service is a well-architected, domain-specialized calculation engine that cleanly separates ephemeris access, mathematical primitives, astrological techniques, and HTTP handling. The package dependency graph is acyclic and the code reads well. However, there are **2 critical** issues (tenant DB isolation, auth fail-open), **6 medium** issues, and several minor improvements that must be addressed before production.

---

## CRITICAL (must fix before deploy)

### C1. Single shared DB pool — no per-tenant isolation
**File:** `/Users/enzo/rag-saldivia/services/astro/cmd/main.go:34,62-69`

The service connects to a **single** `POSTGRES_TENANT_URL` pool and passes it directly to the handler. In a multi-tenant SaaS, this means ALL tenants share one DB connection pool. If two tenants are on different PostgreSQL instances (which is the SDA architecture: PostgreSQL per-tenant), this service can only serve one tenant.

Every query in `queries.sql` correctly filters by `tenant_id`, so there is no cross-tenant data leak **within a single DB instance**. But the architecture requires per-tenant DB resolution via `pkg/tenant.Resolver` for production multi-tenant support.

**Current (dev-only):**
```go
dbURL := config.Env("POSTGRES_TENANT_URL", "")
pool, err = pgxpool.New(ctx, dbURL)
```

**Required for production:**
The handler needs to resolve the correct DB pool per-request using the tenant slug from the JWT context, like other services do (or will need to do). Options:
1. Use `pkg/tenant.Resolver` to get the pool per-request (cleanest, matches SDA pattern)
2. Accept single-tenant mode for dev but document the limitation and add a `TODO` with a tracking issue

**Severity:** CRITICAL for production multi-tenant. Acceptable for single-tenant dev mode if explicitly documented.

### C2. Auth middleware with `FailOpen: true` — security degradation
**File:** `/Users/enzo/rag-saldivia/services/astro/cmd/main.go:89-92`

```go
authMw := sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{
    Blacklist: blacklist,
    FailOpen:  true,
})
```

`FailOpen: true` means that if Redis is unreachable for blacklist checks, the middleware will **allow the request through** even with a potentially revoked token. This is a security risk:
- A revoked token (from logout or password change) will continue to work if Redis is down
- An attacker who causes Redis unavailability can reuse revoked tokens

**Fix:** Change to `FailOpen: false` (the default). The brief Redis unavailability returning 503 is preferable to allowing revoked tokens through. This matches what the other services do (chat, notification, etc. don't set FailOpen).

---

## MEDIUM (must fix before production)

### M1. `http.MaxBytesReader` passes `nil` ResponseWriter
**Files:** `/Users/enzo/rag-saldivia/services/astro/internal/handler/astro.go:100,347,439`

```go
r.Body = http.MaxBytesReader(nil, r.Body, maxBodySize)
```

Every other SDA service passes `w` (the actual ResponseWriter). When `w` is `nil`, `MaxBytesReader` cannot send a 413 response on its own when the limit is exceeded — it will just return an error from `Read()`. While this still works (the `json.Decoder` will get the error), passing `w` is the correct Go pattern and ensures proper HTTP behavior:

```go
r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
```

### M2. No error wrapping with `%w` — impossible to use `errors.Is`
**Files:** All handler and technique code

No Go errors in the service use `%w` for wrapping. The SDA convention is `fmt.Errorf("context: %w", err)`. Currently all errors are created with `fmt.Errorf("msg")` (no wrapping) or returned as bare strings. This means:
- Callers cannot use `errors.Is(err, ErrNotFound)` or `errors.As()`
- Error chains are lost
- The `resolveContact` function returns `http.StatusNotFound` alongside a string error, but the actual `pgx.ErrNoRows` (or equivalent) is silently discarded at line 68

**Fix:** Define sentinel errors and wrap properly:
```go
var ErrContactNotFound = errors.New("contact not found")
// ...
if err != nil {
    return nil, http.StatusNotFound, fmt.Errorf("resolve contact %q: %w", contactName, ErrContactNotFound)
}
```

### M3. Missing `service/` layer — handlers contain business logic
**File:** `/Users/enzo/rag-saldivia/services/astro/internal/handler/astro.go`

The SDA bible specifies a three-layer structure: `handler/ -> service/ -> repository/`. The astro service skips the service layer entirely — handlers directly call technique functions and repository queries. This means:
- Business logic is untestable without HTTP (no unit tests for the orchestration logic)
- The `Query` handler (SSE pipeline) at 70+ lines mixes HTTP, DB, technique orchestration, and LLM calls

**Recommendation:** Extract a `service/astro.go` that owns the orchestration. Handlers become thin HTTP adapters. This is a refactor, not a blocker, but it deviates from SDA convention.

### M4. Dead code: `stub()` function
**File:** `/Users/enzo/rag-saldivia/services/astro/internal/handler/astro.go:465-469`

```go
func stub(w http.ResponseWriter, name string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusNotImplemented)
    json.NewEncoder(w).Encode(map[string]string{"status": "not_implemented", "endpoint": name})
}
```

This function is never called. Leftover from scaffolding. Remove it.

### M5. `Query` endpoint does not validate year range
**File:** `/Users/enzo/rag-saldivia/services/astro/internal/handler/astro.go:339-410`

The `Query` (SSE) endpoint parses its own body struct inline instead of using `parseRequest()`. As a result, it skips the year validation (`< -5000 || > 5000`) that all other endpoints get. An attacker could send `year: 999999` and trigger ephemeris calculations for absurd dates, potentially causing the Swiss Ephemeris to hang or produce garbage.

**Fix:** Either reuse `parseRequest` or add the same validation inline.

### M6. Missing README.md and CHANGELOG.md
**File:** `services/astro/` directory

The SDA bible requires every service to have:
- `README.md` — what it does, endpoints, events
- `CHANGELOG.md` — autogenerated by git-cliff

Neither exists. The `VERSION` file (0.1.0) is present.

---

## LOW (should fix)

### L1. `ListContacts` hardcodes limit=100, offset=0
**File:** `/Users/enzo/rag-saldivia/services/astro/internal/handler/astro.go:424-425`

```go
contacts, err := h.q.ListContacts(r.Context(), repository.ListContactsParams{
    TenantID: tid, UserID: uid, Limit: 100, Offset: 0,
})
```

Pagination is not exposed via query params. For users with many contacts, they cannot page through results. Not critical for a module likely used by individuals, but deviates from the pattern used in other services.

### L2. `CreateContact` returns 409 for all errors — not just conflicts
**File:** `/Users/enzo/rag-saldivia/services/astro/internal/handler/astro.go:453-455`

```go
contact, err := h.q.CreateContact(r.Context(), req)
if err != nil {
    jsonError(w, "create failed", http.StatusConflict)
```

Any DB error (invalid data, connection failure, constraint violation) returns 409 Conflict. Only unique constraint violations should be 409. Other errors should be 500.

### L3. No request ID in error logs
**File:** `/Users/enzo/rag-saldivia/services/astro/internal/handler/astro.go:482`

```go
func sseError(w http.ResponseWriter, f http.Flusher, msg string) {
    slog.Error("astro query error", "error", msg)
```

The SDA convention is to include `middleware.GetReqID(r.Context())` in error logs for request tracing. The `sseError` function doesn't have access to the request. Consider passing the request or at least the request ID.

### L4. `LLM SimplePrompt` is not streaming — fake chunked SSE
**File:** `/Users/enzo/rag-saldivia/services/astro/internal/handler/astro.go:388-401`

```go
response, err := h.llm.SimplePrompt(r.Context(), prompt, 0.7)
// Stream response in ~50 char chunks
for i := 0; i < len(response); i += 50 {
```

The LLM call is synchronous (`SimplePrompt`), then the complete response is chopped into 50-char chunks and sent as SSE events. This gives the appearance of streaming but the client waits for the full LLM response before seeing any tokens. True streaming would use the `Chat` method with streaming support.

This is acceptable for v0.1.0 but should be marked as a known limitation.

### L5. `abs()` function duplicates `math.Abs`
**File:** `/Users/enzo/rag-saldivia/services/astro/internal/technique/solar_return.go:154-159`

```go
func abs(x float64) float64 {
    if x < 0 { return -x }
    return x
}
```

This is a local reimplementation of `math.Abs`. Use `math.Abs` instead.

### L6. `mainPlanets` map duplicated between natal and solar_return
**Files:**
- `/Users/enzo/rag-saldivia/services/astro/internal/natal/chart.go:28-34` — `mainPlanets`
- `/Users/enzo/rag-saldivia/services/astro/internal/technique/solar_return.go:51-57` — `srPlanets`

Both map Spanish names to SwissEph IDs with identical content. Should be a shared constant in `astromath/constants.go` or `ephemeris/`.

### L7. `Eclipses` technique not exposed as an endpoint
The `FindEclipseActivations` function is used internally by `Build()` (the brief) but there is no `/v1/astro/eclipses` endpoint and no corresponding tool in `tools.yaml`. Same for Zodiacal Releasing — computed in the brief but not individually callable via API.

### L8. Zodiacal Releasing endpoint missing
The ZR technique is computed in the context builder but has no dedicated HTTP endpoint. The `tools.yaml` does not list it either. If the agent runtime wants to call ZR independently, it cannot.

---

## INFO (observations, no action needed)

### I1. Architecture — clean dependency graph
```
handler
  -> natal -> ephemeris, astromath
  -> technique -> natal -> ephemeris, astromath
  -> context -> technique, natal, ephemeris, astromath
  -> repository (sqlc generated)
```
No circular dependencies. Package boundaries are clean. `ephemeris` is the single gateway to SwissEph — no other package calls `swephgo` directly.

### I2. Consistent technique pattern
Every technique file follows the same structure:
1. Define result struct with JSON tags
2. Pure function: `(chart, params) -> result`
3. No side effects, no DB access, no HTTP
4. Uses `ephemeris` and `astromath` only

This is excellent. The only outlier is `context/builder.go` which orchestrates all techniques (intentional).

### I3. Test coverage — solid for computation, weak for handlers
| Package | Test file | Coverage |
|---------|-----------|----------|
| ephemeris | sweph_test.go | Good — JulDay, RevJul, CalcPlanet, CalcHouses, EclNut, SolcrossUT |
| astromath | angles_test.go | Good — AngDiff, Normalize360, PosToStr, SignIndex, EclToEq, aspects, bounds |
| natal | chart_test.go | Excellent — golden file comparison against Python |
| technique/solar_arc | solar_arc_test.go | Good — golden + structural |
| technique/primary_dir | primary_dir_test.go | Good — golden + sanity + sphere helpers |
| technique/progressions | progressions_test.go | Good — structural + ingress detection |
| technique/solar_return | solar_return_test.go | Good — golden + lunar returns |
| technique/transits | transits_test.go | Good — golden + structural |
| technique/stations | stations_test.go | Good |
| technique/eclipses | phase10_test.go | Good — eclipses + fixed stars + ZR |
| technique/firdaria | firdaria_test.go | Good — golden + sequence |
| technique/profections | profections_test.go | Good — golden + edge cases |
| context | builder_test.go | Good — full build + section verification |
| handler | astro_test.go | **Weak** — only health, auth-less 401, bad request, year validation |

**Missing tests:**
- Handler tests with mocked auth context (happy path with tenant/user in context)
- SSE Query endpoint tests
- Contact CRUD with mocked DB
- Error handling paths in handlers
- `brief.go` convergence matrix unit tests
- `houses.go` (HouseForLon) — no test file
- `convert.go` (PartOfSpirit, SouthNode, ContraAntiscion) — partially tested

### I4. SDA compliance scorecard

| Criterion | Status | Notes |
|-----------|--------|-------|
| chi router | PASS | Standard middleware chain |
| sqlc for DB | PASS | All queries via sqlc, parameterized |
| slog JSON | PASS | `slog.NewJSONHandler(os.Stdout)` |
| No fmt.Print | PASS | Zero instances |
| JWT auth | PASS | Ed25519 via `pkg/middleware.AuthWithConfig` |
| Header spoofing | PASS | `Auth()` strips headers before JWT verify |
| Tenant from context | PASS | `tenant.FromContext(r.Context())` |
| SQL tenant filter | PASS | All queries filter by `tenant_id + user_id` |
| MaxBytesReader | PARTIAL | Used but passes `nil` (see M1) |
| Health excluded from auth | PASS | Health handler registered outside auth group |
| Error responses generic | PASS | Never exposes stack traces |
| Correct status codes | PARTIAL | 201 for create, but 409 for all errors (see L2) |
| VERSION file | PASS | 0.1.0 |
| README | FAIL | Missing (see M6) |
| CHANGELOG | FAIL | Missing (see M6) |
| service/ layer | FAIL | Missing (see M3) |
| Migrations | PASS | UP + DOWN in `db/tenant/migrations/010_astro_contacts` |
| OpenTelemetry | PASS | `otelhttp.NewHandler`, `sdaotel.Setup` |
| Rate limiting | PASS | 10 req/min per user |
| SecureHeaders | PASS | `sdamw.SecureHeaders()` applied |
| Dockerfile | PASS | Multi-stage, CGO_ENABLED=1, alpine, nobody user, ephe volume |
| NATS events | N/A | This service doesn't publish/consume NATS events (stateless calculator) |

### I5. Primary Directions — O(n^2 * aspects * 2) complexity
`FindDirections` iterates all points x all points x all aspects x 2 (direct + converse). For a chart with ~16 points and 5 aspects, that is ~16 * 16 * 5 * 2 = 2560 iterations, each doing OA calculations. This is fast for a single request but under concurrent load with the 5-minute timeout and CalcMu contention, it could queue up.

The `CalcMu` global mutex is the primary bottleneck: any technique that calls `SetTopo + CalcPlanet` atomically (BuildNatal, Progressions, SolarReturn) holds the lock for the entire calculation. Under load, requests will serialize at this mutex.

### I6. The convergence matrix is a strong differentiator
`brief.go`'s `buildConvergenceMatrix` with technique-weighted scoring per month is a unique feature. The weight hierarchy (Station=4, PD=3, Eclipse/Transit=2, SA=1, LunarReturn=0) is well-considered.

---

## Summary of required actions

| ID | Severity | Action |
|----|----------|--------|
| C1 | CRITICAL | Add per-tenant DB resolution or document single-tenant limitation |
| C2 | CRITICAL | Change `FailOpen: true` to `false` |
| M1 | MEDIUM | Pass `w` to `MaxBytesReader` instead of `nil` |
| M2 | MEDIUM | Add error wrapping with `%w` throughout |
| M3 | MEDIUM | Extract service layer (can be deferred if documented) |
| M4 | MEDIUM | Remove dead `stub()` function |
| M5 | MEDIUM | Add year validation to Query endpoint |
| M6 | MEDIUM | Create README.md and CHANGELOG.md |
| L1-L8 | LOW | Address in follow-up PRs |

**Verdict:** The calculation engine is production-quality. The HTTP/infra layer needs the critical fixes before it can serve multi-tenant traffic safely. Fix C1, C2, M1, M4, M5 and the service is deployable.
