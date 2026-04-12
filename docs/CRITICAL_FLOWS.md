# SDA Framework — Critical Flows

> Last updated: 2026-04-12
> Audience: AI models and engineers operating on this codebase.

This document traces the 5 most critical runtime flows through the SDA Framework,
with exact file paths, function names, line references, NATS subjects, and
invariants. Treat this as the single source of truth for understanding how
requests move through the system.

---

## Table of Contents

1. [Flow 1: Auth (Login, JWT, Refresh, Logout)](#flow-1-auth-login-jwt-refresh-logout)
2. [Flow 2: Multi-Tenant Request Resolution](#flow-2-multi-tenant-request-resolution)
3. [Flow 3: Chat + Agent Pipeline](#flow-3-chat--agent-pipeline)
4. [Flow 4: Document Ingestion Pipeline](#flow-4-document-ingestion-pipeline)
5. [Flow 5: WebSocket Real-Time Flow](#flow-5-websocket-real-time-flow)
6. [Cross-Cutting: NATS Subject Map](#cross-cutting-nats-subject-map)
7. [Cross-Cutting: Tenant Isolation Checklist](#cross-cutting-tenant-isolation-checklist)

---

## Flow 1: Auth (Login, JWT, Refresh, Logout)

### Descripcion

The auth flow covers user authentication from login through JWT issuance,
token refresh (rotation), MFA challenge, and logout with token revocation.
Ed25519 asymmetric signing ensures only the Auth Service can create tokens
while all other services verify with the public key alone.

### Step-by-Step Sequence

```
Step 1: POST /v1/auth/login
  File: services/auth/internal/handler/auth.go:112 (Auth.Login)
  - Reads JSON body {email, password}, validates non-empty
  - Resolves tenant service via resolveService() (line 71)
  - Calls svc.Login()

Step 2: Service Login
  File: services/auth/internal/service/auth.go:123 (Auth.Login)
  - Normalizes email to lowercase (line 125)
  - Queries user via repo.GetUserByEmail (line 128)
  - If user not found: runs bcrypt on dummyHash for timing safety (line 132)
  - Checks is_active + locked_until (lines 145-151)
  - Compares bcrypt hash (line 154)
  - On failure: calls recordFailedLogin -- auto-lockout after 5/20 fails (line 156)
  - On success: calls recordSuccessfulLogin (line 164)

Step 3: MFA Check
  File: services/auth/internal/service/auth.go:167-189
  - Calls CheckMFARequired (line 167)
  - If MFA enabled: issues short-lived MFA token (5min, role="mfa_pending")
  - Returns TokenPair{MFARequired: true, MFAToken: "..."}
  - Client must call POST /v1/auth/mfa/verify to complete login

Step 4: Token Issuance (no MFA or post-MFA)
  File: services/auth/internal/service/auth.go:193-264
  - Fetches user role via getPrimaryRole (line 193)
  - Fetches RBAC permissions via getPermissions (line 199)
  - Creates access token (Ed25519, 15min TTL) with claims: uid, email, name, tid, slug, role, perms
  - Creates refresh token (Ed25519, 7d TTL) with separate JTI
  - Hashes refresh token with SHA-256 (not bcrypt -- JWTs exceed 72-byte limit)
  - Revokes all old refresh tokens for this user (line 234)
  - Stores new refresh hash in DB (line 239)
  - Writes audit log entry (line 249)
  - Publishes NATS event "auth.login_success" (line 255)

Step 5: JWT Creation
  File: pkg/jwt/jwt.go:73 (CreateAccess)
  - Uses Ed25519 signing (gojwt.SigningMethodEdDSA)
  - Auto-generates JTI (UUID) if not provided (line 78)
  - Sets iss="sda", sub=userID, iat, exp, jti
  File: pkg/jwt/jwt.go:96 (CreateRefresh)
  - Same signing, longer TTL (7d default)

Step 6: Response
  File: services/auth/internal/handler/auth.go:161-165
  - Sets HttpOnly secure cookie "sda_refresh" (SameSite=Strict)
  - Returns JSON {access_token, refresh_token, expires_in}

Step 7: Refresh Flow -- POST /v1/auth/refresh
  File: services/auth/internal/handler/auth.go:168 (Auth.Refresh)
  - Reads refresh token from cookie first, falls back to JSON body
  File: services/auth/internal/service/auth.go:358 (Auth.Refresh)
  - Verifies JWT signature + expiry (line 360)
  - Validates SHA-256 hash exists in DB and is not revoked (line 367-375)
  - Revokes old refresh token (single-use rotation, line 379)
  - Re-fetches user data for current role/permissions (lines 382-405)
  - Issues new access + refresh token pair
  - Stores new refresh hash in DB

Step 8: Logout -- POST /v1/auth/logout
  File: services/auth/internal/handler/auth.go:220 (Auth.Logout)
  - Reads refresh token (cookie or body)
  - Extracts access token JTI from Authorization header
  File: services/auth/internal/service/auth.go:459 (Auth.Logout)
  - Revokes refresh token in DB (line 467)
  - Blacklists access token JTI in Redis (line 473)
  File: pkg/security/blacklist.go:34 (TokenBlacklist.Revoke)
  - Stores JTI in Redis with TTL = remaining token lifetime
  - Key format: "sda:token:blacklist:{jti}"

Step 9: Middleware Verification (every authenticated request)
  File: pkg/middleware/auth.go:38 (AuthWithConfig)
  - Strips client-spoofed identity headers X-User-ID, X-Tenant-ID, etc (lines 45-49)
  - Preserves Traefik-injected X-Tenant-Slug (line 42)
  - Extracts Bearer token (line 58)
  - Verifies JWT via sdajwt.Verify (line 64)
  - Checks Redis blacklist if configured (lines 71-87)
  - Rejects MFA-pending tokens (line 90)
  - Injects X-User-ID, X-User-Email, X-User-Role, X-Tenant-ID, X-Tenant-Slug headers
  - Sets tenant.Info in context (line 103)
  - Sets role, permissions, userID, userEmail in context (lines 109-112)
  - Cross-validates JWT slug vs Traefik-injected slug (line 115)
```

### ASCII Diagram

```
Client                    Traefik                Auth Service              Tenant DB         Redis
  |                         |                        |                        |                |
  |-- POST /v1/auth/login ->|                        |                        |                |
  |                         |-- X-Tenant-Slug ------>|                        |                |
  |                         |                        |-- GetUserByEmail ----->|                |
  |                         |                        |<-- {id, hash, ...} ---|                |
  |                         |                        |-- bcrypt.Compare       |                |
  |                         |                        |                        |                |
  |                         |                        |   [if MFA enabled]     |                |
  |                         |                        |-- issue MFA token ---->|                |
  |                         |<-- 200 {mfa_required} -|                        |                |
  |<-- {mfa_token} ---------|                        |                        |                |
  |                         |                        |                        |                |
  |-- POST /v1/auth/mfa/verify ------------------>--|                        |                |
  |                         |                        |-- VerifyTOTP --------->|                |
  |                         |                        |                        |                |
  |                         |                        |-- GetPrimaryRole ----->|                |
  |                         |                        |-- GetPermissions ----->|                |
  |                         |                        |-- Ed25519 sign access  |                |
  |                         |                        |-- Ed25519 sign refresh |                |
  |                         |                        |-- SHA256(refresh) ---->|  (store hash)  |
  |                         |                        |-- audit.Write -------->|                |
  |                         |                        |-- NATS publish ------->|                |
  |                         |<-- 200 {tokens} -------|                        |                |
  |<-- Set-Cookie: sda_refresh (HttpOnly) ----------|                        |                |
  |                         |                        |                        |                |
  |   [Later: authenticated request]                 |                        |                |
  |-- GET /v1/chat/* ------>|                        |                        |                |
  |                         |   [auth middleware]    |                        |                |
  |                         |   Verify(pubKey, jwt)  |                        |                |
  |                         |   IsRevoked(jti) --------------------------------------------->|
  |                         |   <-- false --------------------------------------------------------|
  |                         |   Inject X-User-ID etc |                        |                |
  |                         |-- forward to service ->|                        |                |
  |                         |                        |                        |                |
  |   [Logout]              |                        |                        |                |
  |-- POST /v1/auth/logout ->|                       |                        |                |
  |                         |                        |-- revoke refresh ----->|                |
  |                         |                        |-- blacklist JTI ------------------------------>|
  |                         |<-- 200 logged_out -----|                        |                |
```

### Invariants (MUST NOT Break)

1. **Ed25519 asymmetric signing** -- Only Auth Service has the private key. All other services verify with the public key. A compromised service CANNOT forge tokens.
2. **Refresh token rotation** -- Each refresh token is single-use. On refresh, old token is revoked immediately, new one issued. Prevents replay.
3. **SHA-256 for refresh token hashing** -- NOT bcrypt (JWTs exceed bcrypt's 72-byte truncation limit). High-entropy tokens have no rainbow table risk.
4. **Timing-safe login** -- When user doesn't exist, bcrypt runs on dummyHash (service/auth.go line 132) to prevent email enumeration via response timing.
5. **Cross-tenant validation** -- JWT slug must match Traefik-injected slug (middleware/auth.go line 115). Prevents token replay across tenants.
6. **Header stripping** -- Auth middleware strips ALL identity headers before processing (middleware/auth.go lines 45-49). Client cannot spoof X-User-ID.
7. **MFA token isolation** -- role="mfa_pending" tokens are rejected by auth middleware (middleware/auth.go line 90). Only valid for /v1/auth/mfa/verify.
8. **Token blacklist TTL** -- Blacklisted JTIs expire when the token would have expired naturally. Blacklist doesn't grow forever.
9. **Lockout thresholds** -- 5 failures = 15min lock, 20 failures = permanent lock (admin reset required).

### Common Failure Modes

| Failure | Symptom | Root Cause |
|---------|---------|------------|
| 401 on every request | "invalid token" | JWT_PUBLIC_KEY env var missing or wrong encoding (base64 PEM) |
| Login succeeds but subsequent requests fail | Token created with wrong tenant slug | Traefik not injecting X-Tenant-Slug header |
| Logout doesn't take effect | Revoked token still accepted | Redis down + FailOpen=true, or blacklist not injected into middleware |
| "tenant mismatch" 403 | JWT slug != subdomain slug | User logging in via wrong subdomain, or subdomain routing misconfigured |
| Account locked unexpectedly | ErrAccountLocked | Brute force attempts or automated scanner hitting login endpoint |
| Refresh returns "invalid" | Token already rotated | Client using stale refresh token (race condition with concurrent refresh) |

---

## Flow 2: Multi-Tenant Request Resolution

### Descripcion

Every request in SDA carries a tenant context. Traefik extracts the subdomain
slug, the Auth middleware validates it against the JWT, and the Resolver maps
the slug to the correct PostgreSQL pool and Redis client. Connection pools
are cached per slug to avoid creating a new pool per request.

### Step-by-Step Sequence

```
Step 1: Traefik receives request
  - Extracts subdomain (e.g., saldivia.sda.app -> slug "saldivia")
  - Injects X-Tenant-Slug header
  - Routes to the correct service based on URL path

Step 2: Auth middleware processes JWT
  File: pkg/middleware/auth.go:42 (traefikSlug capture)
  - Saves Traefik slug before stripping headers
  - Verifies JWT -> extracts TenantID and Slug from claims
  - Cross-validates: claims.Slug must equal traefikSlug (line 115)
  - Sets tenant.Info{ID, Slug} in context (line 103)

Step 3: Service resolves tenant DB (if multi-tenant handler)
  File: services/auth/internal/handler/auth.go:71 (resolveService)
  - Reads X-Tenant-Slug from header (line 76)
  - Checks svcCache (sync.Map) for cached service (line 82)
  - If not cached: calls resolver.PostgresPool() and resolver.TenantID()

Step 4: Resolver looks up connection info
  File: pkg/tenant/resolver.go:68 (PostgresPool)
  - Acquires mutex (line 69)
  - Returns cached pool if exists (line 78)
  - Calls createPoolLocked (line 80)

Step 5: Connection info lookup from Platform DB
  File: pkg/tenant/resolver.go:102 (resolveConnInfo)
  - Checks connCache with 5-minute TTL (lines 107-109)
  - Queries platform DB:
    SELECT id, postgres_url, redis_url, postgres_url_enc, redis_url_enc
    FROM tenants WHERE slug = $1 AND enabled = true
  - If encrypted URLs available + encryption key present: decrypts via AES-256 (lines 142-162)
  - Caches result for 5 minutes

Step 6: Pool creation
  File: pkg/tenant/resolver.go:169 (createPoolLocked)
  - Adds sslmode=require if not set (ensureSSLMode, line 175)
  - Releases mutex during network I/O (pool creation) (line 184)
  - Creates pgxpool with MaxConns (default 4 per tenant)
  - Re-acquires mutex, double-checks for race (line 200)
  - Stores in resolver.pools map

Step 7: Redis client creation (similar pattern)
  File: pkg/tenant/resolver.go:211 (createRedisClientLocked)
  - Same pattern: resolve -> parse URL -> create client -> ping -> cache

Step 8: RLS enforcement (when used)
  File: pkg/database/pool.go:54 (SetTenantID)
  - Sets PostgreSQL session variable: SET LOCAL app.tenant_id = $1
  - Enables Row-Level Security policies for tenant isolation

Step 9: Health check loop (background)
  File: pkg/tenant/resolver.go:329 (StartHealthCheck)
  - Goroutine pings all cached pools at interval
  - Unhealthy pools are closed and removed from cache
  - Next request creates a fresh connection
```

### ASCII Diagram

```
Client           Traefik          Auth Middleware       Service Handler       Resolver       Platform DB    Tenant DB
  |                 |                   |                     |                   |                |             |
  | saldivia.sda.app/v1/chat/sessions  |                     |                   |                |             |
  |----->           |                   |                     |                   |                |             |
  |                 | X-Tenant-Slug:    |                     |                   |                |             |
  |                 | "saldivia"        |                     |                   |                |             |
  |                 |-------->          |                     |                   |                |             |
  |                 |                   |                     |                   |                |             |
  |                 |    Verify JWT     |                     |                   |                |             |
  |                 |    claims.Slug == |                     |                   |                |             |
  |                 |    "saldivia" ?   |                     |                   |                |             |
  |                 |    Set context    |                     |                   |                |             |
  |                 |                   |-------->            |                   |                |             |
  |                 |                   |                     |                   |                |             |
  |                 |                   |              resolveService()           |                |             |
  |                 |                   |                     |--- PostgresPool ->|                |             |
  |                 |                   |                     |                   |                |             |
  |                 |                   |                     |         [cache miss]               |             |
  |                 |                   |                     |                   |--- SELECT ---->|             |
  |                 |                   |                     |                   |    tenants     |             |
  |                 |                   |                     |                   |<-- pg_url -----|             |
  |                 |                   |                     |                   |                |             |
  |                 |                   |                     |                   |--- decrypt --->|             |
  |                 |                   |                     |                   |    (AES-256)   |             |
  |                 |                   |                     |                   |                |             |
  |                 |                   |                     |                   |--- pgxpool --->|             |
  |                 |                   |                     |                   |    Create      |             |
  |                 |                   |                     |                   |                |             |
  |                 |                   |                     |<-- pool ----------|                |             |
  |                 |                   |                     |                                    |             |
  |                 |                   |                     |--------- query ------------------->|             |
  |                 |                   |                     |<-------- rows --------------------|             |
  |<--------- JSON response ---------------------------------------------------|                |             |
```

### Invariants (MUST NOT Break)

1. **Slug cross-validation** -- JWT slug must equal Traefik-injected slug. Without this, a valid JWT from tenant A could access tenant B's data.
2. **Pool caching** -- Only one pgxpool is created per tenant slug. Creating a pool per request would exhaust connections.
3. **Mutex release during I/O** -- Resolver releases mutex during pool creation (resolver.go line 184) to avoid blocking all tenants while one pool connects.
4. **Double-check after unlock** -- After re-acquiring mutex, checks if another goroutine created the pool (resolver.go line 200). Prevents duplicate pools.
5. **Connection info cache TTL** -- 5-minute cache (resolver.go line 51). Balances freshness (credential rotation) with performance.
6. **SSL enforcement** -- ensureSSLMode appends sslmode=require if not set (resolver.go line 357). Prevents plaintext DB connections in production.
7. **Enabled check** -- Platform DB query includes `enabled = true`. Disabled tenants get no pool.
8. **MaxConns per tenant** -- Default 4. Prevents one tenant from monopolizing connections.

### Common Failure Modes

| Failure | Symptom | Root Cause |
|---------|---------|------------|
| "unknown tenant" | 502 on all requests for a slug | Tenant not in platform DB or `enabled = false` |
| Connection refused on pool create | 502, logged as "create pool" error | Tenant DB URL wrong or DB down |
| Stale credentials after rotation | Auth works for 5min then fails | connCache TTL (5min) serving old creds |
| All tenants blocked | All requests hang | Mutex not released during pool creation (code bug) |
| Pool exhaustion | Queries timeout | MaxConns too low for tenant's traffic; or connection leak |

---

## Flow 3: Chat + Agent Pipeline

### Descripcion

The chat+agent pipeline handles user messages through two services: Chat
(persistence) and Agent (LLM orchestration). The frontend sends messages
via HTTP or WebSocket mutations. The Agent Runtime runs a loop of
LLM calls + tool executions until it produces a text response.

### Step-by-Step Sequence

```
Step 1: User sends message
  A) Via HTTP: POST /v1/chat/sessions/{sessionID}/messages
     File: services/chat/internal/handler/chat.go:203 (Chat.AddMessage)
  B) Via WebSocket mutation: {type:"mutation", action:"send_message", data:{...}}
     File: services/ws/internal/hub/mutations.go:130 (dispatch -> AddMessage gRPC)

Step 2: Chat handler validates
  File: services/chat/internal/handler/chat.go:203-254
  - MaxBytesReader 1MB (line 204)
  - Validates role in {user, assistant, system} (line 218)
  - Blocks "system" role from external API (line 224) -- only internal services
  - Runs guardrails.ValidateInput on user messages (line 230)
  - Verifies session ownership via GetSession (line 239)

Step 3: Chat service persists message
  File: services/chat/internal/service/chat.go:163 (Chat.AddMessage)
  - Creates message in DB via repo.CreateMessage (line 169)
  - Touches session updated_at (line 182)
  - If role="user": publishes NATS event (line 190)
    Subject: tenant.{slug}.notify.chat.new_message
    Payload: {user_id, type:"chat.new_message", title, body, data:{session_id, message_id}}

Step 4: Frontend calls Agent
  POST /v1/agent/query
  File: services/agent/internal/handler/agent.go:42 (Handler.Query)
  - MaxBytesReader 256KB (line 43)
  - Extracts JWT from Authorization header (line 45)
  - Reads {message, history} (line 48)
  - Calls svc.Query(ctx, jwt, userID, message, history)

Step 5: Agent service runs query loop
  File: services/agent/internal/service/agent.go:82 (Agent.Query)
  - Extracts tenant from context (line 86)
  - Publishes trace start via NATS (line 91)
    Subject: tenant.{slug}.traces.start
  - Runs guardrails.ValidateInput (line 94)
  - Filters history -- only user/assistant roles, truncates content (line 100)
  - Builds message array: [system_prompt, ...history, user_message]

Step 6: Agent loop (LLM -> tools -> LLM -> ...)
  File: services/agent/internal/service/agent.go:118-247
  For each iteration (max 10 by default):
    - Loop detection check (line 119) -- breaks if 3 identical tool calls detected
    - Calls llmAdapter.Chat(ctx, messages, toolSchemas, temperature, maxTokens)
      File: pkg/llm/client.go (Client.Chat)
      - Sends OpenAI-compatible request to SGLang endpoint
    - If no tool calls: applies output guardrails (line 138) -> returns response
    - If tool calls present:
      - For each tool call (max 25 per turn):
        - Validates tool params against JSON schema (line 168)
        - Executes via toolExecutor.Execute(ctx, jwt, toolName, params)
          File: services/agent/internal/tools/executor.go:70
          - If RequiresConfirmation: returns pending_confirmation (line 81)
          - If search_documents + gRPC available: uses gRPC (line 91)
          - Otherwise: HTTP call to target service with JWT forwarding (line 94)
        - Appends tool result to messages for next LLM iteration

Step 7: Tool execution
  File: services/agent/internal/tools/executor.go:114 (executeHTTP)
  - Creates HTTP request to service endpoint
  - Forwards JWT as Authorization: Bearer header (line 120)
  - 30s timeout per tool call
  - Handles 403 as "denied", 4xx+ as "error"

Step 8: Confirmation flow (for destructive tools)
  File: services/agent/internal/handler/agent.go:77 (Handler.Confirm)
  - POST /v1/agent/confirm {tool, params}
  - Calls svc.ExecuteConfirmed(ctx, jwt, tool, params)
  File: services/agent/internal/tools/executor.go:100 (ExecuteConfirmed)
  - Verifies tool has RequiresConfirmation=true (prevents bypass)
  - Executes the tool

Step 9: Trace publishing
  File: services/agent/internal/service/agent.go:266 (publishTraceEnd)
  File: services/agent/internal/service/traces.go (TracePublisher wrapper)
  File: pkg/traces/publisher.go (actual implementation)
  - Publishes trace end event:
    Subject: tenant.{slug}.traces.end
    Payload: {trace_id, status, models_used, duration_ms, tokens, tool_call_count}
  - Publishes feedback event:
    Subject: tenant.{slug}.feedback.usage
  - On non-completed: publishes error report:
    Subject: tenant.{slug}.feedback.error_report

Step 10: Response persisted as assistant message
  - Frontend receives QueryResult from agent
  - Frontend calls POST /v1/chat/sessions/{id}/messages with role="assistant"
  - Chat service persists the response (same flow as Step 3)
```

### ASCII Diagram

```
Frontend         Chat Service          Agent Service          LLM (SGLang)       Tool Services
  |                   |                      |                      |                  |
  |-- POST message -->|                      |                      |                  |
  |   role=user       |                      |                      |                  |
  |                   |-- guardrails ------->|                      |                  |
  |                   |-- CreateMessage      |                      |                  |
  |                   |-- NATS notify ------>|                      |                  |
  |<-- 201 message ---|                      |                      |                  |
  |                   |                      |                      |                  |
  |-- POST /agent/query ------------------>|                      |                  |
  |   {message, history}                    |                      |                  |
  |                   |                      |-- trace.start ------>|                  |
  |                   |                      |-- guardrails         |                  |
  |                   |                      |                      |                  |
  |                   |                      |== AGENT LOOP ========|==================|
  |                   |                      |                      |                  |
  |                   |                      |-- Chat(messages) --->|                  |
  |                   |                      |<-- {tool_calls} -----|                  |
  |                   |                      |                      |                  |
  |                   |                      |-- Execute(tool) ---------------------->|
  |                   |                      |   Authorization: Bearer <jwt>           |
  |                   |                      |<-- {data} -----------------------------|
  |                   |                      |                      |                  |
  |                   |                      |-- Chat(messages     |                  |
  |                   |                      |   + tool_result) -->|                  |
  |                   |                      |<-- {content} -------|                  |
  |                   |                      |                      |                  |
  |                   |                      |== END LOOP =========|==================|
  |                   |                      |                      |                  |
  |                   |                      |-- trace.end -------->|                  |
  |                   |                      |-- feedback.usage --->|                  |
  |<-- 200 QueryResult ---------------------|                      |                  |
  |                   |                      |                      |                  |
  |-- POST message -->|                      |                      |                  |
  |   role=assistant  |                      |                      |                  |
  |                   |-- CreateMessage      |                      |                  |
```

### NATS Subjects in This Flow

| Subject | Publisher | Consumer |
|---------|-----------|----------|
| `tenant.{slug}.notify.chat.new_message` | Chat service | Notification service |
| `tenant.{slug}.traces.start` | Agent service | Traces service |
| `tenant.{slug}.traces.end` | Agent service | Traces service |
| `tenant.{slug}.traces.event` | Agent service | Traces service |
| `tenant.{slug}.feedback.usage` | Agent service | Feedback service |
| `tenant.{slug}.feedback.error_report` | Agent service | Feedback service |

### Invariants (MUST NOT Break)

1. **System role blocked from API** -- Only internal services can set role="system" (chat handler line 224). Prevents prompt injection via message history.
2. **Session ownership** -- GetSession checks userID match before any message operation. User A cannot add messages to User B's session.
3. **Tool confirmation cannot be bypassed** -- ExecuteConfirmed verifies RequiresConfirmation=true (executor.go line 106). Direct call to Confirm on a non-confirmation tool is rejected.
4. **JWT forwarding** -- Agent forwards the user's JWT to tool services (executor.go line 120). Tools inherit the user's RBAC permissions, not a service account.
5. **Loop detection** -- After 3 identical tool calls (same name + params), loop breaks (agent.go line 119). Prevents infinite LLM-to-tool loops.
6. **Max tool calls per turn** -- Hard cap of 25 (agent.go line 157). Prevents runaway cost.
7. **Max loop iterations** -- Hard cap of 10 (default, agent.go line 44). Combined with max tool calls, limits total LLM cost.
8. **Output guardrails** -- System prompt fragments are stripped from LLM output (agent.go line 138). Prevents system prompt leakage.
9. **Input guardrails** -- Both Chat handler and Agent service validate input. Double-layer protection.
10. **History filtering** -- Only user/assistant roles pass through (agent.go line 100). Injected system/tool messages in history are rejected.

### Common Failure Modes

| Failure | Symptom | Root Cause |
|---------|---------|------------|
| Agent returns "query failed" 500 | LLM endpoint unreachable | SGLang not running or AGENT_LLM_ENDPOINT env wrong |
| Tool call always fails | All tools return "error" | Tool service endpoint wrong in tools.yaml, or JWT not forwarded |
| Loop timeout response | "No pude completar la consulta" | LLM keeps calling same tool, loop detection triggers |
| Guardrails blocking valid input | 400 "message blocked" | False positive on block pattern (e.g., Spanish text matching) |
| Missing trace data | Traces service has no records | NATS down or traces consumer not subscribed |
| "pending_confirmation" never resolves | Frontend doesn't call /confirm | Frontend missing confirmation UI flow |

---

## Flow 4: Document Ingestion Pipeline

### Descripcion

The document ingestion pipeline handles file upload, async processing via
NATS JetStream, text extraction (OCR/vision via Extractor service), page
storage, and tree generation. Two parallel paths exist: the legacy Blueprint
path (worker.go) and the modern extraction path (documents.go + extractor_consumer.go).

### Step-by-Step Sequence (Modern Path)

```
Step 1: File upload
  POST /v1/ingest/upload (multipart/form-data)
  File: services/ingest/internal/handler/ingest.go:72 (Ingest.Upload)
  - MaxBytesReader 100MB (line 73)
  - Validates identity headers X-User-ID + X-Tenant-Slug (line 75)
  - ParseMultipartForm with 10MB in-memory buffer (line 81)
  - Validates file extension against allowlist (line 94):
    .pdf, .docx, .doc, .txt, .md, .csv, .xlsx, .pptx, .html, .json, .xml
  - Sanitizes filename -- strips path components (line 100)
  - Requires "collection" form field (line 108)

Step 2: Service stages file + creates job (Legacy/Blueprint path)
  File: services/ingest/internal/service/ingest.go:116 (Ingest.Submit)
  - Creates tenant staging directory: /tmp/ingest-staging/{slug}/ (line 118)
  - Stages file to disk as temp file (line 123-134)
  - Creates job record in DB with status="pending" (line 141)
  - Publishes NATS message for async processing (line 154)
    Subject: tenant.{slug}.ingest.process
    Payload: {job_id, tenant_slug, user_id, collection, file_name, staged_path}
  - If NATS publish fails: deletes staged file + job record (line 170-173)
  - Returns 202 Accepted with job info

Step 2-alt: Service uploads to MinIO + triggers extraction (Modern path)
  File: services/ingest/internal/service/documents.go:50 (DocumentService.UploadDocument)
  - Computes SHA-256 hash via TeeReader (single-pass) (line 54)
  - Dedup check: if hash exists, returns existing document (line 63)
  - Creates document record in DB (line 76)
  - Uploads to MinIO at key: {tenant}/{docID}/original.{ext} (line 89)
  - Publishes extraction job via NATS (line 115)
    Subject: tenant.{slug}.extractor.job
    Payload: {document_id, tenant_slug, storage_key, file_name, file_type}
  - Updates document status to "extracting" (line 125)

Step 3: Worker processes job (Legacy/Blueprint path)
  File: services/ingest/internal/service/worker.go:50 (Worker.Start)
  - Creates JetStream stream "INGEST" (line 57)
    Subjects: tenant.*.ingest.process
    Storage: FileStorage, MaxAge: 24h
  - Creates durable consumer "ingest-worker" (line 67)
    AckPolicy: Explicit, MaxDeliver: 3
  - Fetches 1 message at a time (line 100)

Step 4: Worker processes individual job
  File: services/ingest/internal/service/worker.go:115 (Worker.processJob)
  - Validates tenant from NATS subject matches payload (line 130) -- anti-spoofing
  - Updates job status to "processing" (line 140)
  - Forwards file to Blueprint endpoint POST /v1/documents (line 170)
    - Namespaces collection: {tenant}-{collection} (line 177)
    - Multipart upload with collection_name field
  - On failure:
    - If final attempt (delivery #3): marks "failed", removes file, terminates msg (line 148)
    - Otherwise: Nak for retry, status stays "processing" (line 154)
  - On success: marks "completed", removes staged file, Acks msg (line 160-167)

Step 5: Completion notification
  File: services/ingest/internal/service/worker.go:218 (Worker.publishCompletion)
  - Publishes notification event via EventPublisher.Notify
    Payload: {type:"ingest.completed", user_id, title, body}
  - Publishes WS broadcast for real-time progress
    Subject: tenant.{slug}.ingest.jobs
    Payload: {type:"event", channel:"ingest.jobs", data:{job_id, status, file_name, collection}}

Step 6: Extractor consumer (Modern path -- receives extraction results)
  File: services/ingest/internal/service/extractor_consumer.go:57 (ExtractorConsumer.Start)
  - Creates JetStream stream "EXTRACTOR_RESULTS" (line 65)
    Subjects: tenant.*.extractor.result.>
  - Creates durable consumer "ingest-extractor-consumer" (line 73)
    MaxDeliver: 3, AckWait: 5min

Step 7: Process extraction result
  File: services/ingest/internal/service/extractor_consumer.go:100 (handleResult)
  - Parses ExtractionResult: {document_id, file_name, total_pages, pages[]}
  - Updates document status to "indexing" (line 114)
  - Bulk inserts pages via pgx.CopyFrom (single round-trip) (line 141)
    File: services/ingest/internal/service/bulk_pages.go:23 (BulkInsertPages)
    - Wraps in transaction: DELETE old pages + COPY new pages
  - Generates search tree (line 155)
    - Uses tree.Generator to create hierarchical index
    - Stores tree JSON + doc_description in document_trees table (line 171)
  - Final status: "ready" if all succeeded, "error" if partial failure (line 185-191)
```

### ASCII Diagram

```
Client         Ingest Handler      Ingest Service       MinIO       NATS JetStream     Worker/Consumer    Extractor
  |                  |                    |                 |              |                   |               |
  |-- POST upload -->|                    |                 |              |                   |               |
  |   multipart      |                    |                 |              |                   |               |
  |                  |-- validate ext --->|                 |              |                   |               |
  |                  |                    |-- SHA256 hash   |              |                   |               |
  |                  |                    |-- dedup check   |              |                   |               |
  |                  |                    |-- CreateDoc --->| (DB)         |                   |               |
  |                  |                    |-- Put --------->| (S3)         |                   |               |
  |                  |                    |                 |              |                   |               |
  |                  |                    |-- Publish ----->|              |                   |               |
  |                  |                    |  tenant.{slug}  |              |                   |               |
  |                  |                    |  .extractor.job |              |                   |               |
  |<-- 202 Accepted--|                    |                 |              |                   |               |
  |                  |                    |                 |              |                   |               |
  |                  |                    |                 |   [async]    |                   |               |
  |                  |                    |                 |              |--- Fetch -------->|               |
  |                  |                    |                 |              |                   |-- Extract --->|
  |                  |                    |                 |              |                   |   (OCR/Vision)|
  |                  |                    |                 |              |                   |<-- pages -----|
  |                  |                    |                 |              |                   |               |
  |                  |                    |                 |              |<-- result --------|               |
  |                  |                    |                 |              | tenant.{slug}     |               |
  |                  |                    |                 |              | .extractor.result |               |
  |                  |                    |                 |              |                   |               |
  |                  |                    |                 |   Consumer:  |                   |               |
  |                  |                    |                 |   BulkInsert pages               |               |
  |                  |                    |                 |   Generate tree                  |               |
  |                  |                    |                 |   Status->"ready"                |               |
  |                  |                    |                 |              |                   |               |
  |                  |                    |                 |              |-- WS broadcast -->|               |
  |  [WS: ingest.jobs event with status="completed"]       |              |                   |               |
```

### NATS Subjects in This Flow

| Subject | Publisher | Consumer | Stream |
|---------|-----------|----------|--------|
| `tenant.{slug}.ingest.process` | Ingest service | Worker | INGEST |
| `tenant.{slug}.extractor.job` | DocumentService | Extractor (Python) | -- |
| `tenant.{slug}.extractor.result.>` | Extractor (Python) | ExtractorConsumer | EXTRACTOR_RESULTS |
| `tenant.{slug}.ingest.jobs` | Worker | WS Hub (NATSBridge) | -- |
| `tenant.{slug}.notify.ingest.completed` | Worker (EventPublisher) | Notification service | -- |

### Invariants (MUST NOT Break)

1. **File extension allowlist** -- Only 11 extensions accepted (handler line 26-30). Prevents executable upload.
2. **Filename sanitization** -- filepath.Base strips directory traversal (handler line 100). Prevents path injection.
3. **NATS publish failure = upload failure** -- If NATS publish fails, staged file and DB record are cleaned up (ingest.go lines 170-173). No orphaned pending jobs.
4. **Tenant validation in worker** -- Subject tenant must match payload tenant (worker.go line 130). Prevents cross-tenant data injection.
5. **Dedup by SHA-256** -- Same file content returns existing document, not a duplicate (documents.go line 63).
6. **Bulk insert atomicity** -- Pages are inserted in a single transaction: DELETE old + COPY new (bulk_pages.go lines 28-53). No partial page sets.
7. **3 retries max** -- JetStream MaxDeliver=3. Only final attempt marks job "failed" (worker.go lines 147-154). Earlier failures retry silently.
8. **Collection namespacing** -- Collections are prefixed with tenant slug: `{slug}-{collection}` (worker.go line 177). Prevents cross-tenant collection access.
9. **MaxUploadSize** -- 100MB hard limit (handler line 24). Prevents memory exhaustion.
10. **Tree generation failure is not silent** -- If tree gen fails, document stays in "error" status, not "ready" (extractor_consumer.go line 167-169). Search won't return incomplete documents.
11. **Tenant slug validated for NATS safety** -- DocumentService constructor panics if slug fails `^[a-zA-Z0-9_-]+$` regex (documents.go line 38). Prevents NATS subject injection.

### Common Failure Modes

| Failure | Symptom | Root Cause |
|---------|---------|------------|
| Upload returns 202 but job stays "pending" | Worker not consuming | NATS disconnected or stream not created |
| Job "processing" forever | Worker can't reach Blueprint | BLUEPRINT_URL wrong or Blueprint service down |
| Job "failed" after 3 attempts | Worker logs "blueprint returned HTTP 4xx" | Blueprint rejected file (unsupported format, size, etc.) |
| Duplicate documents | Same file re-uploaded | SHA-256 dedup not enabled (using legacy path) |
| Pages missing from search | BulkInsertPages failed | Transaction rollback -- check DB logs for constraint violations |
| "extractor.job" published but no result | Extractor not running | Python extractor service down or not subscribed to subject |
| MinIO upload fails | UploadDocument returns error | MinIO not reachable or bucket doesn't exist |

---

## Flow 5: WebSocket Real-Time Flow

### Descripcion

The WebSocket Hub is the real-time backbone of SDA. Every live update in the
frontend flows through it. Services publish events to NATS, the NATSBridge
forwards them to connected WebSocket clients filtered by tenant + channel
subscription. Clients can also send mutations (create session, send message)
directly over WebSocket, which are routed to services via gRPC.

### Step-by-Step Sequence

```
Step 1: WebSocket upgrade
  GET /ws with Authorization: Bearer <jwt>
  File: services/ws/internal/handler/ws.go:34 (WS.Upgrade)
  - Extracts JWT from Authorization header only (NOT query param -- avoids log leakage)
  - Verifies JWT via sdajwt.Verify (line 43)
  - Checks token blacklist if configured (line 50)
  - Accepts WebSocket with origin check (line 58-67)
    - WS_ALLOWED_ORIGINS env (comma-separated patterns, e.g. "*.sda.app")
    - If not set: InsecureSkipVerify=true with warning (dev mode)
  - Creates Client with identity from JWT claims (line 74):
    {UserID, Email, TenantID, Slug, Role, JWT}

Step 2: Client registration
  File: services/ws/internal/hub/hub.go:38 (Hub.Run)
  - Hub runs as goroutine with register/unregister channels (line 38)
  - Checks global limit: max 1000 clients (line 44)
  - Checks per-tenant limit: max 300 clients per tenant (line 53)
  - If limit reached: sends error + closes client

Step 3: Auto-subscribe to default channels
  File: services/ws/internal/handler/ws.go:80-82
  - client.Subscribe("notifications")  -- in-app notifications
  - client.Subscribe("modules")        -- module enable/disable events
  - client.Subscribe("presence")       -- user online/offline status

Step 4: Read/Write pumps start
  File: services/ws/internal/hub/client.go:124 (ReadPump -- blocking)
  File: services/ws/internal/hub/client.go:155 (WritePump -- goroutine)
  - ReadPump: reads JSON from WebSocket, dispatches to hub.handleMessage
  - WritePump: reads from send channel (buffered 64), writes to WebSocket
    - 10s write timeout (line 165)
    - If send buffer full: message dropped with warning (TrySend, line 99)

Step 5: Client sends subscribe message
  {type:"subscribe", channel:"chat.messages:session-123"}
  File: services/ws/internal/hub/hub.go:149-165 (handleMessage -> Subscribe)
  - Max 64 subscriptions per client (client.go line 18)
  - Responds with {type:"event", channel:..., data:{subscribed:true}}

Step 6: Client sends mutation
  {type:"mutation", action:"send_message", id:"corr-1", data:{...}}
  File: services/ws/internal/hub/hub.go:176-181 (handleMessage -> Mutation)
  File: services/ws/internal/hub/mutations.go:52 (MutationHandler.Handle)
  - Dispatches in goroutine (non-blocking hub) (line 53)
  - Routes to gRPC ChatService based on action (line 88):
    - "create_session" -> chatClient.CreateSession
    - "delete_session" -> chatClient.DeleteSession
    - "rename_session" -> chatClient.RenameSession
    - "send_message"   -> chatClient.AddMessage
  - Forwards JWT from client to gRPC context (line 91):
    sdagrpc.ForwardJWT(ctx, client.JWT)
  - Sets userID from client identity (e.g., line 99: req.UserId = client.UserID)
  - On success: sends {type:"event", id:"corr-1", data:{...}} back
  - On error: sends {type:"error", id:"corr-1", error:"...", data:{code:"token_expired"}}
  - Token expiry detected: code "token_expired" (mutations.go line 62)

Step 7: NATS -> WebSocket bridge
  File: services/ws/internal/hub/nats.go:31 (NATSBridge.Start)
  - Subscribes to "tenant.*.>" (all tenant events) (line 33)
  File: services/ws/internal/hub/nats.go:52 (handleNATSMessage)
  - Parses subject: tenant.{slug}.{channel}
  - Extracts tenantSlug from parts[1], channel from parts[2] (lines 54-61)
  - Parses NATS payload as WebSocket Message
  - If not a valid Message: wraps raw data as Event (line 67)
  - Calls hub.BroadcastToTenant(tenantSlug, channel, wsMsg) (line 82)

Step 8: Hub broadcasts to subscribed clients
  File: services/ws/internal/hub/hub.go:110 (BroadcastToTenant)
  - Iterates all clients under RLock
  - Filters by: client.Slug == tenantSlug AND client.IsSubscribed(channel)
  - Sends via TrySend (non-blocking, drops if buffer full)

Step 9: Client disconnects
  File: services/ws/internal/hub/client.go:124-128 (ReadPump defer)
  - Sends self to hub.unregister channel
  - Closes WebSocket connection
  File: services/ws/internal/hub/hub.go:70-82 (Hub.Run -> unregister)
  - Removes client from clients map
  - Calls markClosed() -- atomic CAS + close(send) (client.go lines 117-120)
```

### ASCII Diagram

```
Frontend (WS)          WS Hub              NATSBridge           NATS             Chat (gRPC)
     |                    |                     |                  |                   |
     |-- GET /ws -------->|                     |                  |                   |
     |   Bearer <jwt>     |                     |                  |                   |
     |                    |-- Verify JWT        |                  |                   |
     |                    |-- Register client   |                  |                   |
     |                    |-- Auto-subscribe:   |                  |                   |
     |                    |   notifications     |                  |                   |
     |                    |   modules           |                  |                   |
     |                    |   presence          |                  |                   |
     |<-- connected ------|                     |                  |                   |
     |                    |                     |                  |                   |
     |== SUBSCRIBE =======|                     |                  |                   |
     |{type:"subscribe",  |                     |                  |                   |
     | channel:"chat.     |                     |                  |                   |
     | messages:abc-123"} |                     |                  |                   |
     |<-- {subscribed:    |                     |                  |                   |
     |     true} ---------|                     |                  |                   |
     |                    |                     |                  |                   |
     |== MUTATION ========|                     |                  |                   |
     |{type:"mutation",   |                     |                  |                   |
     | action:            |                     |                  |                   |
     | "send_message",    |                     |                  |                   |
     | id:"c1",           |                     |                  |                   |
     | data:{...}}        |                     |                  |                   |
     |                    |-- gRPC AddMessage ---------------------------------->|
     |                    |   ForwardJWT(jwt)    |                  |                   |
     |                    |<-- response ------------------------------------------------|
     |<-- {type:"event",  |                     |                  |                   |
     |     id:"c1",       |                     |                  |                   |
     |     data:{msg}}    |                     |                  |                   |
     |                    |                     |                  |                   |
     |== SERVER PUSH =====|                     |                  |                   |
     |                    |                     |<-- msg ----------|                   |
     |                    |                     |  tenant.saldivia |                   |
     |                    |                     |  .chat.messages  |                   |
     |                    |                     |                  |                   |
     |                    |<-- BroadcastToTenant|                  |                   |
     |                    |  slug=saldivia      |                  |                   |
     |                    |  channel=           |                  |                   |
     |                    |  chat.messages      |                  |                   |
     |                    |                     |                  |                   |
     |<-- {type:"event",  |                     |                  |                   |
     |  channel:          |                     |                  |                   |
     |  "chat.messages",  |                     |                  |                   |
     |  data:{...}} ------|                     |                  |                   |
```

### WebSocket Protocol Reference

```
File: services/ws/internal/hub/protocol.go

Message types:
  "subscribe"   -- client -> server: subscribe to channel updates
  "unsubscribe" -- client -> server: unsubscribe from channel
  "mutation"    -- client -> server: execute an action (create, update, delete)
  "event"       -- server -> client: data push or mutation response
  "error"       -- server -> client: error message

Message envelope:
  {
    "type":    "<MessageType>",
    "channel": "chat.messages:session-123",  // for subscribe/event
    "action":  "send_message",               // for mutations
    "id":      "corr-1",                     // correlation ID (request/response)
    "data":    {...},                         // payload
    "error":   "something went wrong"        // only for error type
  }

Standard channels (hub/protocol.go:35-46):
  "sessions"          -- session list updates
  "chat.messages"     -- chat messages (append :session_id for specific session)
  "notifications"     -- in-app notifications (auto-subscribed)
  "admin.stats"       -- admin dashboard stats
  "ingest.jobs"       -- document ingestion progress
  "presence"          -- user online/offline (auto-subscribed)
  "collections"       -- collection updates
  "modules"           -- module enable/disable (auto-subscribed)
  "fleet.vehicles"    -- fleet module: vehicle updates
  "fleet.maintenance" -- fleet module: maintenance events

Mutation actions (hub/mutations.go:93-145):
  "create_session"  -> Chat.CreateSession gRPC
  "delete_session"  -> Chat.DeleteSession gRPC
  "rename_session"  -> Chat.RenameSession gRPC
  "send_message"    -> Chat.AddMessage gRPC
```

### Invariants (MUST NOT Break)

1. **JWT in header only** -- Token is extracted from Authorization header, NEVER from query parameters (ws.go line 36). Prevents JWT leakage in access logs.
2. **Blacklist check on connect** -- Revoked tokens are rejected at WebSocket upgrade (ws.go line 50). Prevents logged-out users from maintaining connections.
3. **Global + per-tenant connection limits** -- Max 1000 total, max 300 per tenant (hub.go lines 10-11). Prevents DoS.
4. **Send buffer overflow = drop, not block** -- TrySend returns false if buffer full (client.go line 99). Hub never blocks on slow clients.
5. **Atomic close** -- markClosed uses atomic CAS (client.go line 118). Prevents write-after-close panic on send channel.
6. **Mutations run in goroutine** -- Handle dispatches asynchronously (mutations.go line 53). Hub event loop never blocks on gRPC calls.
7. **UserID from JWT, not from client** -- Mutation handler overrides req.UserId with client.UserID (mutations.go lines 99, 115, 122, 136). Client cannot impersonate.
8. **Tenant isolation in broadcast** -- BroadcastToTenant checks client.Slug == tenantSlug (hub.go line 121). Tenant A's events never reach Tenant B's clients.
9. **Max 64 subscriptions per client** -- Prevents resource abuse (client.go line 18).
10. **NATSBridge wildcard** -- Subscribes to "tenant.*.>" to catch ALL tenant events. Any new NATS subject automatically flows to WebSocket without code changes.
11. **JWT forwarded on mutations** -- gRPC calls include the original user JWT (mutations.go line 91). Services verify the token independently.

### Common Failure Modes

| Failure | Symptom | Root Cause |
|---------|---------|------------|
| WebSocket connects but no events arrive | Client subscribed, NATS events published | NATSBridge not started or NATS connection lost |
| "server at capacity" on connect | Hub reached MaxClients | Connection leak -- clients not being unregistered on disconnect |
| "tenant connection limit reached" | Hub reached MaxClientsPerTenant | Too many tabs/devices for one tenant, or connection leak |
| Mutation returns "token_expired" | gRPC returns Unauthenticated | JWT expired during long WebSocket session -- client must refresh and reconnect |
| Messages dropped silently | slog.Warn "send buffer full" | Client is slow consumer -- increase sendBufSize or investigate client-side read speed |
| "mutations not available" | MutationHandler is nil | CHAT_GRPC_TARGET env not set or chat gRPC server unreachable |
| Events from wrong tenant | Security breach | BroadcastToTenant slug filter bypassed (verify hub.go line 121) |
| Stale data after reconnect | No events while disconnected | WebSocket is live-only -- no message replay. Client must HTTP-fetch on reconnect. |

---

## Cross-Cutting NATS Subject Map

All NATS subjects follow the pattern `tenant.{slug}.{service}.{event}`:

```
tenant.{slug}.notify.{eventType}       -> Notification Service (JetStream)
tenant.{slug}.ingest.process           -> Ingest Worker (JetStream: INGEST)
tenant.{slug}.ingest.jobs              -> WS Hub (via NATSBridge)
tenant.{slug}.extractor.job            -> Extractor Service (Python)
tenant.{slug}.extractor.result.>       -> ExtractorConsumer (JetStream: EXTRACTOR_RESULTS)
tenant.{slug}.traces.start             -> Traces Service
tenant.{slug}.traces.end               -> Traces Service
tenant.{slug}.traces.event             -> Traces Service
tenant.{slug}.feedback.{category}      -> Feedback Service
tenant.{slug}.chat.messages            -> WS Hub (via NATSBridge)
tenant.{slug}.sessions                 -> WS Hub (via NATSBridge)
tenant.{slug}.notifications            -> WS Hub (via NATSBridge)
tenant.{slug}.modules                  -> WS Hub (via NATSBridge)
tenant.{slug}.presence                 -> WS Hub (via NATSBridge)
tenant.{slug}.collections              -> WS Hub (via NATSBridge)
```

Slug validation: `^[a-zA-Z0-9_-]+$` enforced in:
- `pkg/nats/publisher.go:19` (IsValidSubjectToken)
- `pkg/traces/publisher.go:15` (ValidateToken)
- `services/ingest/internal/service/documents.go:24` (safeSubjectToken)

Every publisher validates the slug before constructing the subject.

---

## Cross-Cutting: Tenant Isolation Checklist

For any new flow or modification, verify these properties hold:

- [ ] JWT slug cross-validated against Traefik slug (pkg/middleware/auth.go:115)
- [ ] Identity headers stripped before processing (pkg/middleware/auth.go:45-49)
- [ ] NATS subject includes tenant slug; consumer validates it
- [ ] Database queries scoped to tenant (either per-tenant DB or RLS)
- [ ] MinIO/S3 keys prefixed with tenant slug
- [ ] Redis keys namespaced per tenant
- [ ] WebSocket broadcasts filtered by tenant slug
- [ ] Tool calls forward user JWT (not service account)
- [ ] Collections namespaced: `{slug}-{collection}`
- [ ] Slug validated against `^[a-zA-Z0-9_-]+$` before any NATS publish
