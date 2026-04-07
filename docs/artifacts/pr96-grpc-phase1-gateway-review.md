# Gateway Review -- PR #96 Plan 09 Phase 1 (gRPC Foundation)

**Fecha:** 2026-04-05
**PR:** #96 (2 commits: dcc64a3, c932490)
**Branch:** `feat/plan09-phase-1` -> `2.0.1` (post-merge review)
**Resultado:** APROBADO con correcciones menores para follow-up

---

## Scope

Phase 1 of Plan 09 delivers the gRPC foundation:
1. **buf config** (`proto/buf.yaml`, `proto/buf.gen.yaml`)
2. **Proto updates** (fixed `go_package` in 6 existing protos, added `search.proto`)
3. **Generated code** (14 `.pb.go` files in `gen/go/`, new `gen/go/go.mod`)
4. **`pkg/grpc`** (interceptors, server factory, client factory)
5. **Infrastructure** (`go.work` updated, `Makefile` proto target fixed)

---

## Bloqueantes

None. All 7 compile-error blockers and 8 parity gaps from the plan review (plan09-grpc-gateway-review.md) are resolved.

---

## Debe corregirse (follow-up, not blocking)

### D1. Interceptor does not inject UserID/Email into context
**[pkg/grpc/interceptors.go:102-108]**

The HTTP middleware (`pkg/middleware/auth.go:96-99`) injects identity via headers:
```go
r.Header.Set("X-User-ID", claims.UserID)
r.Header.Set("X-User-Email", claims.Email)
r.Header.Set("X-User-Role", claims.Role)
```

The gRPC interceptor injects `tenant.Info`, `role`, and `permissions` into the context -- but **not** the user identity (`UserID`, `Email`, `Name`). Every HTTP handler in the codebase reads `r.Header.Get("X-User-ID")` (grep finds 30+ call sites across chat, notification, ingest, auth, feedback, platform).

When future phases implement gRPC handlers for chat, ingest, or notification, they will need UserID from the context and it won't be there.

**Fix:** Add `WithUserID(ctx, claims.UserID)` and `WithUserEmail(ctx, claims.Email)` helpers to `pkg/middleware/rbac.go` (or a new `pkg/middleware/identity.go`), and call them in the interceptor alongside `WithRole`/`WithPermissions`. This also benefits HTTP handlers that could migrate from header-reading to context-reading over time.

**Priority:** Must be done before Phase 4 (agent -> search doesn't need it; search is tenant-scoped. But Phase 5 agent -> ingest will need UserID).

### D2. No tests for `pkg/grpc`
**[pkg/grpc/]**

The plan listed `pkg/grpc/grpc_test.go` as a deliverable. The package has 3 files and 0 tests. The interceptor logic (JWT extraction, blacklist check, MFA rejection, context injection) is critical security code that should have unit tests.

**Fix:** Add `pkg/grpc/interceptors_test.go` covering:
- Valid JWT -> context has tenant, role, permissions
- Missing metadata -> codes.Unauthenticated
- Invalid JWT -> codes.Unauthenticated
- Revoked token -> codes.Unauthenticated
- MFA-pending role -> codes.Unauthenticated
- Health check method -> bypasses auth
- FailOpen=true with Redis error -> passes
- FailOpen=false with Redis error -> codes.Unavailable

### D3. `rag.proto` references deprecated service
**[proto/rag/v1/rag.proto]**

The RAG service is deprecated (replaced by agent). The proto file header says "Thin proxy to the NVIDIA RAG Blueprint" which is no longer accurate. The generated code for it (`gen/go/rag/v1/`) compiles and is harmless, but the comments are misleading.

**Fix:** Either remove `proto/rag/v1/rag.proto` entirely (since no gRPC handler will ever implement it), or add a comment marking it as deprecated. The plan itself notes this proto is "No (deprecated)" for gRPC migration.

---

## Sugerencias

### S1. Health service registration helper in `NewServer`
The interceptor correctly skips `/grpc.health.v1.Health/Check` for auth, but `NewServer()` doesn't register the standard gRPC health service. Consider adding a `NewServerWithHealth()` variant or documenting that callers should register `grpc_health_v1.RegisterHealthServer()` themselves. This becomes important when Docker Compose or future k8s uses gRPC health probes.

### S2. Consider validating `UnaryInterceptor` not in caller opts
`NewServer()` uses `append(defaults, opts...)` which means a caller passing `grpc.UnaryInterceptor()` (singular) would silently replace the auth chain. The doc comment warns about this, which is good. A defensive option would be to validate that no `UnaryInterceptor` option is in `opts`, or to put auth interceptors last (after caller opts) so they always run. Not critical since all callers are internal.

### S3. buf.yaml lint exceptions could use per-rule comments
The 4 lint exceptions in `buf.yaml` have a single comment explaining `PACKAGE_DIRECTORY_MATCH`. The other 3 (`RPC_REQUEST_RESPONSE_UNIQUE`, `RPC_REQUEST_STANDARD_NAME`, `RPC_RESPONSE_STANDARD_NAME`) are all justified by the existing proto design (e.g., `CreateSession` returns `Session` not `CreateSessionResponse`), but each could benefit from a one-line comment for future maintainers.

### S4. Consider `optional` field sentinel for SearchRequest.max_nodes
`SearchRequest.max_nodes` is `int32` (default 0). A caller that doesn't set it will get 0, which downstream code should treat as "use default" rather than "return 0 results". This is a proto design choice, not a bug, but the handler must handle the zero case.

---

## Lo que esta bien

### Interceptor parity with HTTP middleware
All 7 compile-error blockers from the plan review are fixed:
- `sdajwt.Verify(cfg.PublicKey, token)` -- correct arg order (was reversed in plan draft)
- `tenant.WithInfo()` -- correct function (plan draft used nonexistent `sdajwt.NewContext()` and `tenant.NewContext()`)
- `claims.Slug` -- correct field name (was `claims.TenantSlug`)
- `grpc.NewClient()` -- uses non-deprecated API (not `grpc.DialContext`)

All 5 parity gaps from the plan review are addressed:
- Role + permissions injection via `sdamw.WithRole()` and `sdamw.WithPermissions()`
- Blacklist check with `FailOpen` semantics matching HTTP
- MFA-pending rejection
- Generic error messages (no internal detail leakage)
- `JWTFromIncomingContext()` helper for gRPC-to-gRPC chaining

### buf configuration
- `buf.yaml` v2 syntax is correct
- `buf.gen.yaml` uses remote plugins (no local protoc required)
- `paths=source_relative` correctly places output in `gen/go/{service}/v1/`
- Lint exceptions are all justified by existing proto design
- Breaking change detection enabled (`FILE` policy)

### Generated code
- 14 `.pb.go` files generated for 7 protos (7 message + 7 gRPC stubs)
- `gen/go/go.mod` properly declares the module with grpc v1.80.0 and protobuf v1.36.11
- `go.work` includes `./gen/go`
- Package names follow convention (`authv1`, `chatv1`, `searchv1`, etc.)

### search.proto design
- Pulled forward from Phase 4 into Phase 1 (correct -- codegen is the right place)
- Clean unary RPC design matching existing HTTP handler
- `Selection` message captures all tree search output (document, nodes, sections, pages, text, tables, images)
- `optional collection_id` correctly uses proto3 optional

### Client factory
- `grpc.NewClient()` (not deprecated `DialContext`)
- Insecure transport with clear documentation ("internal Docker network only")
- Keepalive params match the plan's security design
- `ForwardJWT()` correctly re-adds "Bearer " prefix after `JWTFromIncomingContext()` strips it

### Server factory
- `ChainUnaryInterceptor` / `ChainStreamInterceptor` (not the singular interceptor that replaces)
- 4MB default recv limit
- Keepalive params match plan specification
- Comment warns about the `UnaryInterceptor` trap

### Security
- No JWT token logged (only error reason)
- Error messages to client are generic ("invalid token", "missing authorization")
- Traefik cross-validation intentionally omitted (correct for internal-only gRPC)
- `wrappedStream` correctly overrides `Context()` for stream interceptors

### Makefile
- `proto` target runs `buf lint && buf generate` then `go mod tidy` in gen/go
- Clean output messaging

---

## Checklist against plan review blockers

| Plan review blocker | Status |
|---|---|
| B1: `sdajwt.NewContext()` does not exist | FIXED -- uses `tenant.WithInfo()` |
| B2: `tenant.NewContext()` does not exist | FIXED -- uses `tenant.WithInfo()` |
| B3: `claims.TenantSlug` does not exist | FIXED -- uses `claims.Slug` |
| B4: `sdajwt.Verify(token, pub)` arg order | FIXED -- `Verify(cfg.PublicKey, token)` |
| B5: `grpc.DialContext` deprecated | FIXED -- uses `grpc.NewClient` |
| B6: `h.svc.Query()` does not exist | N/A (Phase 4, not in this PR) |
| B7: `derefString()`/`toProto()` missing | N/A (Phase 4, not in this PR) |
| D1: Missing role/permissions injection | FIXED |
| D2: No blacklist check | FIXED |
| D3: No MFA-pending rejection | FIXED |
| D4: Error message leaks | FIXED -- generic messages only |
| D5: No `JWTFromIncomingContext` | FIXED |
