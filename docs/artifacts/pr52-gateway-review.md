# Gateway Review -- PR #52 feat/full-wiring

**Fecha:** 2026-04-03
**Resultado:** CAMBIOS REQUERIDOS

**Scope:** Auth Service (new endpoints), Auth Middleware, CLI, MCP Server, Traefik configs, go.work, module manifests, OpenAPI specs.

---

## Bloqueantes

### B1. refresh_tokens.id scanned as int, column is TEXT
**File:** `services/auth/internal/service/auth.go:215`

The `refresh_tokens` table defines `id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text` (migration line 57), but the Refresh method scans the id into `var id int`:

```go
var id int
err = a.db.QueryRow(ctx,
    `SELECT id FROM refresh_tokens
     WHERE token_hash = $1 AND user_id = $2 AND revoked_at IS NULL AND expires_at > now()`,
    tokenHash, claims.UserID,
).Scan(&id)
```

pgx will return a scan error because it cannot scan a UUID string into an int. **Every refresh request will fail with an internal error.** This means users cannot refresh their tokens, which makes the entire auth flow broken after the first access token expires (15 minutes).

**Fix:** Change `var id int` to `var id string`. Or better, since the id value is never used, use a simple existence check:

```go
var exists bool
err = a.db.QueryRow(ctx,
    `SELECT EXISTS(SELECT 1 FROM refresh_tokens
     WHERE token_hash = $1 AND user_id = $2 AND revoked_at IS NULL AND expires_at > now())`,
    tokenHash, claims.UserID,
).Scan(&exists)
```

---

### B2. Refresh endpoint does not return new refresh_token in body for CLI/MCP clients
**File:** `services/auth/internal/handler/auth.go:126-129`

The Login handler returns the full `TokenPair` (including `refresh_token` in body, line 84), allowing CLI/MCP clients to extract both tokens. However, the Refresh handler only returns `access_token` + `expires_in`:

```go
writeJSON(w, http.StatusOK, map[string]any{
    "access_token": tokens.AccessToken,
    "expires_in":   tokens.ExpiresIn,
})
```

The new refresh token is only set as an HttpOnly cookie. Non-browser clients (CLI, MCP) that sent the old refresh token via JSON body will:
1. Have their old token revoked (rotation happened, line 229).
2. Receive a new access token.
3. Have NO way to get the new refresh token (no cookie jar).
4. Be unable to refresh again when the new access token expires.

This effectively gives CLI/MCP clients a single-use refresh before they must login again.

**Fix:** Include `refresh_token` in the response body, matching the Login endpoint behavior:

```go
writeJSON(w, http.StatusOK, map[string]any{
    "access_token":  tokens.AccessToken,
    "refresh_token": tokens.RefreshToken,
    "expires_in":    tokens.ExpiresIn,
})
```

---

### B3. Deploy command hardcodes docker-compose.dev.yml for production deploys
**File:** `tools/cli/cmd/deploy.go:50-51`

The deploy command is described as "Deploy a service to production" but uses the dev compose file:

```go
{"Pulling latest image", "docker", []string{"compose", "-f", "deploy/docker-compose.dev.yml", "pull", service}},
{"Restarting service", "docker", []string{"compose", "-f", "deploy/docker-compose.dev.yml", "up", "-d", "--no-deps", service}},
```

Running `sda deploy auth` in production would pull and restart using dev configuration. The bible says "Deploy a prod es siempre manual: `sda deploy auth`" -- this command needs to target the production compose file.

**Fix:** Either accept a `--env` flag (defaulting to `prod`), or use environment detection, or have a separate compose file path config:

```go
composeFile := env("SDA_COMPOSE_FILE", "deploy/docker-compose.prod.yml")
```

---

## Debe corregirse

### D1. CLI service health decodes body after resp.Body.Close()
**File:** `tools/cli/cmd/service.go:64-65`

```go
resp.Body.Close()

var body map[string]string
if resp.StatusCode == http.StatusOK {
    json.NewDecoder(resp.Body).Decode(&body)
```

The body is closed on line 62, then decoded on line 65. The decoder will read from a closed body, which returns an error silently. The `body` variable is never used anyway (the status is determined by HTTP status code, not body content), but this is a correctness bug that will cause confusing behavior if body parsing is added later.

**Fix:** Move `resp.Body.Close()` to after the decode, or use `defer resp.Body.Close()` before the decode block:

```go
resp, err := client.Get(url)
if err != nil {
    fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", svc.Name, svc.Port, "DOWN", "-")
    continue
}
defer resp.Body.Close()
```

Wait, `defer` in a loop is a leak. Better:

```go
func checkHealth(client *http.Client, url string) (int, error) {
    resp, err := client.Get(url)
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()
    io.Copy(io.Discard, resp.Body) // drain
    return resp.StatusCode, nil
}
```

---

### D2. Traefik prod tenant-from-subdomain uses rewriteHeaders plugin (not built-in)
**File:** `deploy/traefik/dynamic/prod.yml:93-99`

The `rewriteHeaders` is a Traefik **plugin**, not a built-in middleware. It requires explicit installation in the Traefik static config (via `experimental.plugins.rewriteHeaders`). If the plugin is not installed, Traefik will reject the dynamic config silently and the `X-Tenant-Slug` header will never be set.

This is not just a deployment issue -- it is a **security concern**: if the plugin fails to load, requests will arrive at services without `X-Tenant-Slug`, and services that rely on it for tenant resolution may fall back to a default or error. Make sure the static Traefik config properly declares this plugin.

Additionally, the regex `^([^.]+)\.sda\.app$` does not sanitize the captured group. A subdomain like `admin--injection.sda.app` would set `X-Tenant-Slug: admin--injection`. While the Resolver's SQL query uses parameterized queries (safe), it is worth adding a validation step downstream or constraining the regex to `^([a-z0-9-]+)\.sda\.app$` to match only valid slugs.

**Fix:**
1. Verify the static Traefik config declares the `rewriteHeaders` plugin.
2. Tighten the regex: `"^([a-z0-9][a-z0-9-]{0,62})\\.sda\\.app$"`

---

### D3. Traefik prod does NOT strip pre-existing X-Tenant-Slug headers
**File:** `deploy/traefik/dynamic/prod.yml`

The `tenant-from-subdomain` middleware sets `X-Tenant-Slug` from the Host header, but it does **not strip** a pre-existing `X-Tenant-Slug` header from the client request. If the plugin fails to match (e.g., the Host header doesn't match the regex), the original client-supplied header could pass through.

In the dev config, `dev-tenant` middleware sets the header unconditionally via `customRequestHeaders`, which overwrites any existing value (correct). But in prod, the behavior depends on the plugin -- if the regex doesn't match, does it clear the header or leave it?

**Fix:** Add a middleware before `tenant-from-subdomain` that explicitly strips `X-Tenant-Slug`:

```yaml
strip-tenant-headers:
  headers:
    customRequestHeaders:
      X-Tenant-Slug: ""
```

Then chain: `[strip-tenant-headers, tenant-from-subdomain, rate-limit, cors]`

---

### D4. Auth service is hardcoded to a single tenant (not multi-tenant ready)
**File:** `services/auth/cmd/main.go:31-32`

```go
tenantID := env("TENANT_ID", "dev")
tenantSlug := env("TENANT_SLUG", "dev")
```

The auth service creates a single `service.Auth` instance bound to one tenant. This works for dev with a single tenant, but the spec says Auth Gateway is a core service handling all tenants. In production, each subdomain routes to the same auth service, but the service only knows about one tenant's database.

This is a design limitation, not a bug per se for the current dev phase. But it means deploying to production requires either:
- Running one auth service instance per tenant (expensive, doesn't match the spec).
- Refactoring to use the `tenant.Resolver` to look up the correct DB per request.

**Recommendation:** Document this as a known limitation. Add a TODO comment in main.go: `// TODO: Refactor to use tenant.Resolver for multi-tenant support`.

---

### D5. OpenAPI spec error responses don't match handler format
**File:** `docs/api/auth.yaml` components vs `services/auth/internal/handler/auth.go`

The OpenAPI spec defines `ErrorResponse` with two fields:
```yaml
ErrorResponse:
  properties:
    error:    # machine-readable code ("invalid_credentials")
    message:  # human-readable message
```

But the actual handler uses a single `error` field:
```go
type errorResponse struct {
    Error string `json:"error"`
}
// Usage:
writeJSON(w, http.StatusBadRequest, errorResponse{Error: "email and password are required"})
```

The spec says `error` contains a machine-readable code like `"invalid_credentials"` and `message` contains the human string. The handler puts the human string in `error` and has no `message` field.

**Fix:** Either update the handler to match the spec:
```go
type errorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
}
```

Or update the OpenAPI spec to match the handler (single `error` field with the message). Pick one and be consistent.

---

### D6. OpenAPI spec cookie name mismatch
**File:** `docs/api/auth.yaml:51`

The spec shows:
```
Set-Cookie: refresh_token=eyJ...; Path=/; HttpOnly; Secure; SameSite=Strict
```

But the actual handler uses `sda_refresh` as the cookie name:
```go
http.SetCookie(w, &http.Cookie{
    Name: "sda_refresh",
```

**Fix:** Update the OpenAPI spec examples to use `sda_refresh`.

---

### D7. OpenAPI EnabledModule.id is UUID format, handler returns string
**File:** `docs/api/auth.yaml:351-352` vs `services/auth/internal/handler/auth.go:186-201`

The spec defines:
```yaml
EnabledModule:
  properties:
    id:
      type: string
      format: uuid
```

But the handler returns plain string IDs like `"chat"`, `"rag"`:
```go
{ID: "chat", Name: "Chat", Category: "core"},
```

Also, the spec enum for category is `[core, ai, business]` but the manifests use `core` and `vertical`. The spec should match reality.

**Fix:** Remove `format: uuid` from the `id` property. Update `category` enum to `[core, platform, vertical, ai_service]` to match the plan's module categories.

---

### D8. Module manifest: ingest marked as core, spec says it is modular
**File:** `modules/ingest/manifest.yaml:3`

```yaml
category: core
```

The spec at line 146 explicitly lists Ingest as a **modular** service, not core: "Ingest -- Upload de documentos, pipeline de procesamiento via NATS, conectores (Google Drive, OneDrive, etc.)." The core services are: Auth, WS Hub, Chat, RAG, Notification, Platform.

The EnabledModules handler also hardcodes ingest as always-returned, which contradicts the modular nature.

**Fix:** Change `category: core` to `category: platform` in the manifest. Remove ingest from the hardcoded core modules list in the handler (it should come from Platform DB once integrated).

---

### D9. Logout does not verify token before revoking (minor security)
**File:** `services/auth/internal/service/auth.go:299-309`

The Logout method only hashes the token and updates the DB. It does not verify the JWT signature. This means any arbitrary string that happens to SHA-256-hash to a stored value could revoke a token. In practice the probability is negligible (SHA-256 collision), but for consistency with Refresh (which verifies the JWT first), Logout should also verify.

More importantly, the Logout handler in `auth.go:133-155` does not require authentication (it is NOT in the protected route group in `cmd/main.go:98`). Anyone can POST to `/v1/auth/logout` with a token in the body. The revocation only affects that specific token so the blast radius is limited, but it allows an attacker with a captured token to force-logout the user.

**Fix:** Consider moving `/v1/auth/logout` into the authenticated route group, or at least verify the JWT signature in the Logout service method.

---

## Sugerencias

### S1. MCP Server db_query tool is defined but not implemented
**File:** `tools/mcp/main.go`

The `db_query` tool is defined in the tool list (line 84-100) and promises "Only SELECT queries are allowed", but the `handleToolCall` switch only handles `service_health` and `tenant_list` (both returning TODO stubs). The `db_query` and `service_logs` tools will fall through to the default case and return "unknown tool" errors.

This is not a bug (tools are stubs), but the tool definitions should either be removed until implemented or the handler should return a clear "not yet implemented" message for all defined tools. Currently it is confusing because the tool shows up in the MCP tools list but errors when called.

### S2. CLI tenant list uses hardcoded port 8006
**File:** `tools/cli/cmd/tenant.go:45`

```go
url := fmt.Sprintf("http://%s:8006/v1/platform/tenants", baseHost)
```

Should use the `PLATFORM_PORT` env var or at least match the configurable pattern used elsewhere.

### S3. CLI uses `http://` for all connections
**Files:** `tools/cli/cmd/service.go:52`, `tools/cli/cmd/tenant.go:45`

In production, connections should be over TLS. The CLI should support a `SDA_SCHEME` or `SDA_URL` env var, or default to HTTPS.

### S4. Refresh cookie path is `/v1/auth` -- consider implications
**File:** `services/auth/internal/handler/auth.go:213`

The cookie path `Path: "/v1/auth"` means the browser will only send the cookie on requests to `/v1/auth/*`. This is correct security practice (minimal scope). However, ensure the frontend's refresh logic targets `/v1/auth/refresh` and not a different path.

### S5. Consider adding CORS origin for the cookie to work cross-origin
**File:** `deploy/traefik/dynamic/prod.yml:122-126`

The CORS config includes `accessControlAllowCredentials: true`, which is necessary for cookies. Good. The origin regex `^https://[a-z0-9-]+\\.sda\\.app$` is appropriately restrictive.

### S6. Auth handler enabledModules should read X-Tenant-ID to filter modules
**File:** `services/auth/internal/handler/auth.go:185-201`

The EnabledModules handler is protected by auth middleware (reads JWT), but it does not use the tenant context at all. The TODO comment acknowledges this. When implementing, it should use the tenant ID from context to query the Platform DB for that tenant's enabled modules.

### S7. NATS publisher Notify validates tenant slug but not event type
**File:** `pkg/nats/publisher.go:57`

The subject is built as `"tenant." + tenantSlug + ".notify." + parsed.Type`. The slug is validated via `isValidSubjectToken`, but `parsed.Type` is not. An event type like `foo.bar>` would create an invalid NATS subject. Should validate the type too.

**Fix:**
```go
if !isValidSubjectToken(parsed.Type) {
    return fmt.Errorf("invalid event type for NATS subject: %q", parsed.Type)
}
```

Note: This was partially identified in PR #34 review (NATS channel validation gap in Broadcast) and the Broadcast method was fixed. The Notify method has the same pattern but the event type is not validated.

### S8. Race condition in resolver createPoolLocked
**File:** `pkg/tenant/resolver.go:124-154`

The double-check-after-unlock pattern is correct, but there is a subtle issue: if two goroutines resolve the same slug concurrently, both will call `resolveConnInfo` (which checks the cache under the lock), then both release the lock to create pools. The first one back wins and stores the pool; the second detects the existing pool and closes its own. This works correctly but wastes a connection during the race. The code correctly handles this and is safe.

### S9. Auth service should log on successful startup with tenant info
**File:** `services/auth/cmd/main.go:114`

The startup log says `"auth service starting"` but does not log which tenant it is serving. Add `"tenant_id", tenantID, "tenant_slug", tenantSlug` to help with debugging multi-instance deployments.

### S10. Consider rate limiting the refresh endpoint
The login endpoint has brute-force protection (account lockout). The refresh endpoint has no rate limiting beyond the Traefik global rate limit. An attacker with a stolen refresh token could hammer the endpoint to generate access tokens before the original user notices. The token rotation helps (each refresh invalidates the old token), but consider adding per-user rate limiting.

---

## Lo que esta bien

### Auth Service
- Timing-safe login: `dummyHash` comparison when user doesn't exist prevents user enumeration via response timing. Properly applied for both non-existent and disabled users.
- Bcrypt cost 12 is appropriate.
- Refresh token rotation: old token revoked before new one issued, single-use design.
- SHA-256 for refresh token hashing is correct -- JWTs exceed bcrypt's 72-byte limit, and high-entropy tokens don't need slow hashing.
- Error wrapping follows `fmt.Errorf("context: %w", err)` pattern consistently.
- Account lockout after 5 failed attempts with 15-minute lockout window.
- Login response does not leak account state (disabled accounts return same error as wrong password).
- All SQL uses parameterized queries (`$1`, `$2`), no string interpolation.
- Audit log captures user, action, IP, and user agent.
- NATS event publishing is fire-and-forget (logged on error, doesn't block the response).
- Access tokens: 15min, Refresh: 7d -- matches spec.

### Auth Middleware (pkg/middleware/auth.go)
- Strips spoofed headers (`X-User-ID`, `X-User-Email`, `X-User-Role`, `X-Tenant-ID`, `X-Tenant-Slug`) BEFORE processing JWT. Correct anti-spoofing pattern.
- Excludes `/health` from auth check.
- Sets `tenant.Info` in context for downstream services.
- JWT Verify rejects `alg: none` via `SigningMethodHMAC` type assertion in `pkg/jwt/jwt.go:91`.
- JWT secret minimum 32 bytes enforced in `CreateAccess`/`CreateRefresh`.
- JWT claims include all required fields: uid, email, name, tid, slug, role.
- `Verify` checks for empty `UserID`, `TenantID`, `Slug` (ErrMissingClaim).

### Cookie Security
- `HttpOnly: true` -- not accessible via JavaScript.
- `Secure: true` -- only sent over HTTPS.
- `SameSite: http.SameSiteStrictMode` -- strongest CSRF protection.
- Cookie path scoped to `/v1/auth` -- minimal exposure.
- Clear on logout sets `MaxAge: -1`.

### NATS Publisher
- `isValidSubjectToken` rejects NATS special characters (`.`, `*`, `>`, whitespace).
- Broadcast method validates both slug AND channel (fix from previous review).
- Subject format follows spec: `tenant.{slug}.notify.{type}` and `tenant.{slug}.{channel}`.

### Tenant Resolver
- Connection pooling with lazy creation and caching.
- Double-check-after-unlock pattern prevents duplicate pool creation.
- 5-minute TTL on connection info cache.
- Platform DB query uses parameterized queries.
- Close() properly cleans up all pools and clients.

### CLI
- Service allowlist in deploy prevents arbitrary service names.
- Dry-run flag for deploy previews.
- Token required for tenant operations (SDA_TOKEN env var).
- Version command supports build-time injection via ldflags.

### MCP Server
- Clean JSON-RPC implementation following MCP protocol.
- Proper error codes (-32601 method not found, -32602 invalid params).
- 1MB buffer limit on stdin scanner.
- Logging goes to stderr (correct for MCP -- stdout is the protocol channel).

### Traefik Dev Config
- `dev-tenant` middleware unconditionally sets `X-Tenant-ID` and `X-Tenant-Slug`, overwriting any spoofed values.
- Services correctly point to `host.docker.internal` for host development.

### Migrations
- UP and DOWN migrations present.
- DOWN drops tables in correct dependency order.
- Proper use of `IF NOT EXISTS` / `IF EXISTS`.
- System roles seeded with `ON CONFLICT DO NOTHING`.
- Indexes on frequently queried columns (user_id, token_hash, created_at).

### go.work
- Both `tools/cli` and `tools/mcp` properly added to workspace.
- All 7 services + 2 tools + pkg registered.

### Module Manifests
- Consistent structure: id, name, category, tier_min, icon, routes, nav, api_endpoints, ws_channels.
- Vertical modules (fleet, construction) declare `requires` dependencies.
- Navigation positions don't conflict (0, 10, 20 for core; 40, 41 for verticals).

---

## Summary of changes required

| Priority | Count | Items |
|----------|-------|-------|
| Bloqueantes | 3 | B1 (type mismatch crashes refresh), B2 (CLI refresh broken), B3 (deploys hit dev) |
| Debe corregirse | 9 | D1-D9 |
| Sugerencias | 10 | S1-S10 |

The bloqueantes must be fixed before merge. B1 alone breaks the entire refresh flow.
