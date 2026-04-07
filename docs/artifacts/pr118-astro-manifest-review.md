# Gateway Review -- PR #118 Astro Tool Manifest + Docker Compose (Phase 15)

**Fecha:** 2026-04-06
**Resultado:** CAMBIOS REQUERIDOS

---

## Bloqueantes

### B1. Docker Compose: missing `REDIS_URL` env var

`services/astro/cmd/main.go:56` calls `security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))`. The fallback `localhost:6379` works when running on host (`make dev`), but in Docker (`make dev-full`) the Redis container is named `redis`, not `localhost`. Every other service gets Redis connectivity through `*env-common` or explicit env, but blacklist init will silently connect to nothing and JWT blacklist checks will fail open.

**Fix:** Add `REDIS_URL: redis:6379` to the astro environment block in `docker-compose.dev.yml`.

### B2. Docker Compose: missing `SGLANG_LLM_URL` env var

`services/astro/cmd/main.go:35-36` reads `SGLANG_LLM_URL` and `SGLANG_LLM_MODEL`. Without these, the `/v1/astro/query` SSE endpoint falls back to sending the raw brief (no LLM narration). The agent service at line 307 sets `SGLANG_LLM_URL: http://host.docker.internal:8102` -- astro should do the same if it wants LLM narration in Docker mode.

**Fix:** Add to astro environment:
```yaml
SGLANG_LLM_URL: http://host.docker.internal:8102
SGLANG_LLM_MODEL: ${SGLANG_LLM_MODEL:-}
LLM_API_KEY: ${LLM_API_KEY:-}
```

### B3. Traefik dev.yml: astro router missing

`deploy/traefik/dynamic/dev.yml` has no entry for astro. When running with `make dev` (Go on host, Traefik in Docker), requests to `/v1/astro/*` will return 404. Only Docker labels (profile `full`) are configured.

**Fix:** Add to `dev.yml`:
```yaml
# routers section:
astro:
  rule: "PathPrefix(`/v1/astro`)"
  service: astro
  entryPoints: [web]
  middlewares: [dev-cors, dev-tenant]

# services section:
astro:
  loadBalancer:
    servers:
      - url: "http://host.docker.internal:8011"
```

---

## Debe corregirse

### D1. tools.yaml: `contacts` endpoints missing from manifest

The handler exposes `GET /v1/astro/contacts` and `POST /v1/astro/contacts` (lines 117-118 of main.go), but `tools.yaml` does not declare them. If the Agent Runtime uses the manifest to discover available tools, contact management will be invisible. Either add `list_contacts` + `create_contact` tools or document that contacts are managed outside the agent tool system.

### D2. tools.yaml: protocol mismatch with fleet pattern

Fleet tools use `protocol: grpc` with a `method:` field. Astro tools use `protocol: http` with an `endpoint:` field. This is not wrong per se (astro is HTTP, fleet is gRPC), but the Agent Runtime tool executor must handle both protocols. Verify the Agent Runtime dispatcher supports `endpoint` + `protocol: http`. If it only speaks gRPC to module services, all astro tools are unreachable.

### D3. tools.yaml: `astro_query` is SSE, not JSON

The `astro_query` tool (line 165) points to `POST /v1/astro/query` which returns `text/event-stream`, not `application/json`. The Agent Runtime needs to know this is a streaming endpoint. Consider adding a `response_format: sse` or `streaming: true` field to distinguish it from the JSON tools. Without this, the executor may try to parse SSE as JSON and fail.

### D4. `astro_ephe` volume: empty named volume

`docker-compose.dev.yml` mounts `astro_ephe:/ephe:ro` but this is a Docker named volume that is never populated. The Swiss Ephemeris data files must be present at `/ephe` or the service falls back to Moshier (lower precision). Either:
- Add a comment explaining Moshier fallback is intentional for dev, or
- Add an init step or bind-mount pointing to local ephemeris files.

---

## Sugerencias

- **Port comment in dev.yml:** Update the ports comment block at the top of `dev.yml` to include `astro:8011`. Currently the last port listed is `search:8010`.
- **Year validation:** `tools.yaml` does not declare `year` constraints. The handler rejects `year < -5000 || year > 5000` -- consider adding `minimum`/`maximum` to the YAML schema so the Agent Runtime can validate before calling.
- **`requires_confirmation: false` for `intelligence_brief`:** This tool runs ALL techniques, which is computationally expensive. Consider `requires_confirmation: true` or at minimum documenting the cost.
- **Rate limit:** The rate limit of 10 req/min/user is applied to all endpoints including the lightweight ones (profections, firdaria). The brief and query endpoints are the expensive ones. Consider a separate rate limit group.

---

## Lo que esta bien

- **Endpoint-to-manifest 1:1 match:** All 10 technique endpoints in `main.go` (natal, transits, solar-arc, directions, progressions, returns, profections, firdaria, fixed-stars, brief) plus query have corresponding tool entries in `tools.yaml`.
- **Tenant isolation correct:** `resolveContact` uses `tenant.FromContext` + `UserIDFromContext`, and the sqlc query filters by both `tenant_id` AND `user_id`.
- **MaxBytesReader used consistently:** Both `parseRequest` and `CreateContact` and `Query` limit body to 1MB.
- **Auth middleware applied:** `sdamw.AuthWithConfig(publicKey, ...)` protects all routes; `/health` is outside the auth group.
- **Docker Compose follows existing pattern:** `*service-defaults`, `*env-common`, profile `full`, Traefik labels with priority 100 -- all consistent with other services.
- **SSE endpoint correctly excludes chi timeout:** Query route is in its own group without `middleware.Timeout`, relying on `WriteTimeout: 5m` on the HTTP server.
