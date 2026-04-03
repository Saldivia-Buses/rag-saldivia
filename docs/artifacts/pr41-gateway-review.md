# Gateway Review -- PR #41 feat/dev-stack-routing

**Fecha:** 2026-04-02
**Tipo:** review
**Intensity:** thorough
**Reviewer:** gateway-reviewer agent
**Branch:** feat/dev-stack-routing -> 2.0.x
**Scope:** docker-compose.dev.yml, traefik dynamic config, Makefile, chat + notification integration tests

---

## Resultado

**CAMBIOS REQUERIDOS** (2 bloqueantes, 4 correcciones, rest sugerencias)

---

## Hallazgos

### Bloqueantes

1. **[docker-compose.dev.yml:139] ws service sets `WS_ALLOWED_ORIGINS: "*"`**

   The ws service sets `WS_ALLOWED_ORIGINS: "*"` which disables origin checking. The handler at `services/ws/internal/handler/ws.go:53` logs a warning when this is set to wildcard but still accepts all origins. In `dev-full` mode (all services in Docker), someone could connect a WebSocket from any origin. This is technically acceptable for local development, but the compose file has no comment marking it as dev-only, and copy-pasting this block to a production compose is a real risk.

   **Fix:** Add a comment `# DEV ONLY — production must use explicit origins` and consider using `http://localhost:3000,http://localhost:80` even in dev to catch CORS issues early. The wildcard means real CORS bugs won't surface until prod.

2. **[docker-compose.dev.yml:222-224] platform service has JWT_SECRET hardcoded inline, not via anchor**

   The `platform` service sets `JWT_SECRET: dev-secret-at-least-32-characters-long!!` as a plain string instead of using `<<: *env-common`. Same for `ws` (line 139). This creates drift risk -- if someone changes the secret in the anchor, `platform` and `ws` will have a different secret and JWT verification will silently fail. In a multi-service JWT system, all services MUST share the exact same signing secret.

   **Fix:** Either use `<<: *env-common` for both services (and override only the keys they need), or extract `JWT_SECRET` into a proper `.env` file that all services reference. For `ws`:
   ```yaml
   ws:
     <<: *service-defaults
     environment:
       <<: *env-common
       WS_PORT: "8002"
       WS_ALLOWED_ORIGINS: "http://localhost:3000"  # DEV ONLY
   ```
   For `platform`:
   ```yaml
   platform:
     # ...
     environment:
       <<: *env-common
       PLATFORM_PORT: "8006"
       POSTGRES_PLATFORM_URL: postgres://sda:sda_dev@postgres:5432/sda_platform?sslmode=disable
   ```
   This way `JWT_SECRET` and `NATS_URL` are always consistent across services.

### Debe corregirse

3. **[docker-compose.dev.yml:169-186] rag service missing `<<: *service-defaults` and has no depends_on**

   The `rag` service defines `build` and `restart` manually instead of using the YAML anchor `<<: *service-defaults`. While `rag` doesn't need postgres/redis/nats (it only proxies to the NVIDIA Blueprint), the inconsistency makes the compose harder to read. More importantly, if rag gains a NATS dependency in the future (likely, per spec: "Registra la query en audit log del tenant"), the missing anchor will be a silent bug.

   **Fix:** Use `<<: *service-defaults` and add a comment explaining that rag's only external dependency is the Blueprint. If you want to avoid the unnecessary health-check wait, at minimum add the anchor and override `depends_on: {}`.

4. **[docker-compose.dev.yml:215-235] platform service missing `<<: *service-defaults`**

   Same as rag. Platform does depend on postgres (correctly listed), but bypasses the anchor. It also does not depend on `redis` or `nats` (correct -- platform doesn't use them today). But it doesn't even get `restart: unless-stopped` from the anchor... wait, it does set it manually (line 220). Still, for consistency, use the anchor.

   **Fix:** Use `<<: *service-defaults` and override `depends_on` to only include postgres (same pattern as notification).

5. **[docker-compose.dev.yml:112-128] auth service missing TENANT_SLUG env var**

   Auth's `cmd/main.go` reads `TENANT_SLUG` (line 31) with fallback `"dev"`. Chat explicitly sets `TENANT_SLUG: dev` (line 159). Auth doesn't set it. In dev the fallback works, but this inconsistency will bite when someone tries to test multi-tenant by changing the slug -- they'll change it in chat but forget auth. Tenant slug mismatch between auth and chat means the JWT claims won't match the chat service's context.

   **Fix:** Add `TENANT_SLUG: dev` and `TENANT_ID: dev` to the `&env-common` anchor or to the auth service's environment block. Consider adding both to `&env-common` since most tenant-scoped services need them:
   ```yaml
   x-env-common: &env-common
     NATS_URL: nats://nats:4222
     JWT_SECRET: dev-secret-at-least-32-characters-long!!
     POSTGRES_TENANT_URL: postgres://sda:sda_dev@postgres:5432/sda_tenant_dev?sslmode=disable
     TENANT_SLUG: dev
     TENANT_ID: dev
   ```

6. **[docker-compose.dev.yml:237-243] ingest service missing JWT_SECRET explicitly (relies on &env-common)**

   This is actually correct -- ingest inherits `JWT_SECRET` from `&env-common`. No fix needed. Noting for completeness that this is the right pattern and other services (ws, platform) should follow it.

### Sugerencias

7. **[docker-compose.dev.yml:26] JWT_SECRET is a plain env var, bible says "never in env vars planos"**

   The bible (`docs/bible.md:201`) states: "Secrets en Docker secrets o Vault, nunca en env vars planos". This compose file has `JWT_SECRET` as a plain environment variable. For development this is acceptable and pragmatic, but should be documented as a deliberate deviation. The production compose MUST use Docker secrets.

   **Suggestion:** Add a file-level comment at the top: `# DEVELOPMENT ONLY — secrets are plain env vars. Production uses Docker secrets.`

8. **[docker-compose.dev.yml:33-39] postgres credentials are hardcoded**

   `POSTGRES_USER: sda` and `POSTGRES_PASSWORD: sda_dev` are fine for dev. Just noting this is expected for development and should not carry over to production compose. The current inline comment structure makes the dev intent clear enough.

9. **[traefik/dynamic/dev.yml:51-55] platform route has no dev-tenant middleware**

   The platform service route does NOT include `dev-tenant` middleware. This is correct -- platform operates on the platform DB, not tenant DBs. Platform admins use a different auth flow and don't have a tenant context. The comment "admin-only tenant/module/config management" confirms intent.

10. **[traefik/dynamic/dev.yml:17-20] auth route has no dev-tenant middleware**

    Also correct. Auth doesn't need tenant context for login -- it resolves the tenant from the JWT claims or the request itself. Good design.

11. **[traefik/dynamic/dev.yml] No rate limiting middleware in dev**

    The dev Traefik config has no rate limiting. This means integration testing won't catch rate-limiting-related bugs. Consider adding a generous rate limit middleware (e.g., 100 req/s) to match production topology even in dev.

12. **[Makefile:27-33] dev/dev-full/stop targets are clean and correct**

    The `-d` (detached) flag is deliberately omitted from `make dev` and `make dev-full`, so logs stream to terminal. This is the right UX for development. `make stop` uses `--profile full` to ensure all containers (including profiled services) are stopped.

    **Minor suggestion:** Consider adding `make dev-d` (detached) variant for when developers want to run services in background:
    ```makefile
    dev-d: ## Start infra only (detached)
    	docker compose -f $(DEPLOY_DIR)/docker-compose.dev.yml up -d
    ```

13. **[chat_integration_test.go:198] AddMessage for "assistant" role ignores error**

    Line 198: `svc.AddMessage(ctx, session.ID, "u-1", "assistant", "Hola! ...", nil, nil)` -- the returned error is silently discarded. If this insert fails, the subsequent assertion on `len(messages) == 2` will fail with a confusing error message instead of a clear "add message failed" message.

    **Fix:** Capture and check the error:
    ```go
    _, err = svc.AddMessage(ctx, session.ID, "u-1", "assistant", "Hola! En que puedo ayudarte?", nil, nil)
    if err != nil {
        t.Fatalf("add assistant message: %v", err)
    }
    ```

14. **[chat_integration_test.go:149,169,243-246] Multiple CreateSession/DeleteSession calls ignore errors**

    Several test helper calls discard errors (e.g., lines 149, 169, 243-246). If the precondition setup fails, the test will produce misleading failures. Use `t.Helper()` patterns or at minimum check setup errors.

15. **[notification_integration_test.go:131-133, 185-187] Same pattern -- Create calls discard errors**

    Same as chat tests. Setup calls to `svc.Create(...)` discard errors.

16. **[chat_integration_test.go + notification_integration_test.go] Each test function spins up a new postgres container**

    Every test function calls `setupTestDB(t)` which spins up a fresh testcontainers postgres. This is correct for isolation but slow (each container takes 3-5s to start). This matches the pattern established in auth's integration tests, so it's consistent.

    **Future optimization:** Use `TestMain` to spin up one container and use transactions for isolation (begin tx before test, rollback after). This would cut test time from ~45s to ~5s for the suite. Not a blocker for this PR.

17. **[chat_integration_test.go] Missing test: AddMessage to session owned by another user (service layer)**

    The handler layer test (`chat_test.go:409`) covers the ownership check before AddMessage, but the service-level integration test doesn't verify that `AddMessage` at the service layer allows cross-user message insertion. This is by design (service layer trusts the caller), but documenting this trust boundary with a comment in the test would help.

18. **[notification_integration_test.go:163-176] TestMarkRead_WrongUser tests error but not error type**

    The test checks `err == nil` but doesn't assert the specific error. Per the service code, `MarkRead` for a wrong user returns `ErrNotificationNotFound` (the service deliberately doesn't distinguish "not yours" from "doesn't exist" to prevent enumeration). The test should verify the specific error:
    ```go
    if !errors.Is(err, ErrNotificationNotFound) {
        t.Fatalf("expected ErrNotificationNotFound, got: %v", err)
    }
    ```

### Lo que esta bien

- **YAML anchor design (`x-service-defaults`, `x-env-common`)** is excellent. Reduces duplication, makes the compose file maintainable. The pattern of overriding `depends_on` for notification (which needs mailpit) is clean.

- **Traefik dual-provider priority system** is well-designed. File provider routes (auto-priority ~14) are always overridden by Docker label routes (explicit priority 100) in `dev-full` mode. In `dev` mode (services on host), only file routes exist. No conflicts. The comment at the top of the compose file explains both modes clearly.

- **dev-tenant middleware** correctly injects `X-Tenant-ID` and `X-Tenant-Slug` headers for all tenant-scoped services, and correctly omits them for auth and platform. This matches the spec: "In production, Traefik extracts these from the subdomain."

- **Service port allocation** is clean and sequential: auth:8001, ws:8002, chat:8003, rag:8004, notification:8005, platform:8006, ingest:8007. No collisions, easy to remember.

- **Integration test schema setup** in both test files correctly creates the minimal schema needed (users table for FK + service-specific tables). The schemas match the real migrations closely enough to be useful.

- **Test isolation via per-test containers** guarantees no state leakage between tests. Tests run independently and can be parallelized in the future.

- **Ownership verification tests** are present in both services:
  - Chat: `TestGetSession_Ownership_Integration` verifies ErrNotOwner
  - Notification: `TestMarkRead_WrongUser_Integration` verifies cross-user rejection

- **Cascade delete test** (`TestDeleteSession_CascadesMessages_Integration`) verifies the FK cascade at the database level, which is exactly the kind of thing that unit tests with mocks can't catch.

- **Notification preferences test** (`TestPreferences_DefaultsAndUpdate_Integration`) verifies the full UPSERT cycle including default values on first read, which exercises the `ON CONFLICT DO UPDATE` SQL.

- **Build tag `//go:build integration`** correctly gates these tests behind `-tags=integration`, preventing them from running in CI unit test suites that don't have Docker.

- **Makefile targets** are consistent with the compose file: `make dev` = infra only, `make dev-full` = everything, `make stop` = stop all. The `test-integration` target correctly passes `-tags=integration`.

- **Notification service depends_on override** is correct: it lists postgres, nats, and mailpit (which it needs) and drops redis (which it doesn't use). This is better than the default anchor which would force waiting for redis.

---

## Summary

The PR is solid infrastructure work. The two bloqueantes are about JWT secret consistency across services (duplicated inline values vs. anchor reference) and the wildcard CORS origin. Both are quick fixes. The integration tests follow the established pattern from auth and test meaningful scenarios. The Traefik routing is clean and the dual-provider priority system is well-thought-out.

After fixing items 1-2 (bloqueantes) and 3-5 (corrections), this is ready to merge.
