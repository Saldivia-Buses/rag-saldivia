# Gateway Review -- Plan 09: gRPC Inter-Service Communication

**Fecha:** 2026-04-05
**Resultado:** CAMBIOS REQUERIDOS

**Plan:** `/home/enzo/rag-saldivia/docs/plans/2.0.x-plan09-grpc.md`
**Archivos verificados:** 6 proto files, `pkg/jwt/jwt.go`, `pkg/tenant/context.go`, `pkg/middleware/auth.go`, `pkg/middleware/rbac.go`, `services/agent/internal/tools/executor.go`, `services/agent/cmd/main.go`, `services/search/cmd/main.go`, `services/search/internal/service/search.go`, `services/search/internal/handler/search.go`

---

## Bloqueantes

### B1. `sdajwt.NewContext(ctx, claims)` does not exist [interceptors.go:377]

The interceptor code at line 377 calls `sdajwt.NewContext(ctx, claims)`. This function does not exist anywhere in `pkg/jwt/`. The JWT package has no context helpers. The HTTP middleware in `pkg/middleware/auth.go` does NOT inject claims into context via a jwt function -- it injects them as individual pieces:

- `tenant.WithInfo(ctx, ...)` for tenant context
- `middleware.WithRole(ctx, ...)` for role
- `middleware.WithPermissions(ctx, ...)` for permissions

**Fix:** Remove the `sdajwt.NewContext` call. Instead, replicate what the HTTP middleware does:

```go
ctx = tenant.WithInfo(ctx, tenant.Info{
    ID:   claims.TenantID,
    Slug: claims.Slug,
})
ctx = sdamw.WithRole(ctx, claims.Role)
ctx = sdamw.WithPermissions(ctx, claims.Permissions)
```

This requires importing `pkg/middleware` into `pkg/grpc`, which creates a circular dependency risk. Better approach: extract `WithRole`/`WithPermissions` into `pkg/jwt` or a new `pkg/identity` package, or have `pkg/grpc` only inject tenant info and let each gRPC handler do its own permission check.

---

### B2. `tenant.NewContext` does not exist [interceptors.go:372]

The interceptor uses `tenant.NewContext(ctx, tenant.Info{...})`. The actual function is `tenant.WithInfo(ctx, info)`.

**Fix:** Replace `tenant.NewContext` with `tenant.WithInfo`.

---

### B3. `claims.TenantSlug` does not exist [interceptors.go:374]

The interceptor references `claims.TenantSlug`. The actual field in `pkg/jwt.Claims` is `claims.Slug` (line 37 of jwt.go). `TenantSlug` would fail to compile.

**Fix:** Change `claims.TenantSlug` to `claims.Slug`.

---

### B4. `grpc.DialContext` is deprecated in grpc v1.79.2 [client.go:451]

The plan uses `grpc.DialContext(ctx, target, ...)`. This function was deprecated in grpc-go v1.63 (January 2024). The codebase already has grpc v1.79.2 as an indirect dependency. Using deprecated API will cause lint warnings and is bad practice in new code.

**Fix:** Use `grpc.NewClient(target, ...)` instead of `grpc.DialContext`. Note that `grpc.NewClient` does NOT take a context (it returns a lazy connection). The signature change also means that `insecure.NewCredentials()` is the correct transport credential (already used), but error handling changes -- `grpc.NewClient` never returns an error for DNS resolution; it fails lazily on first RPC. Update the `Dial` function accordingly:

```go
func Dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
    defaults := []grpc.DialOption{
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithKeepaliveParams(keepalive.ClientParameters{...}),
        grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(DefaultMaxRecvMsgSize)),
    }
    return grpc.NewClient(target, append(defaults, opts...)...)
}
```

This also changes the call sites since `Dial` no longer takes `ctx`.

---

### B5. `h.svc.Query(...)` -- wrong method name [handler/grpc.go:562]

The search gRPC handler calls `h.svc.Query(ctx, ...)`. The actual service method is `h.svc.SearchDocuments(ctx, ...)` (line 62 of `services/search/internal/service/search.go`). This would fail to compile.

**Fix:** Change to `h.svc.SearchDocuments(ctx, req.Query, derefString(req.CollectionId), int(req.MaxNodes))`.

---

### B6. `derefString` is referenced but never defined [handler/grpc.go:562]

The gRPC handler calls `derefString(req.CollectionId)` but the function is never defined anywhere in the plan. `req.CollectionId` is `*string` (proto3 `optional string`), so it needs dereferencing, but the helper must be provided.

**Fix:** Add the helper function to the plan:

```go
func derefString(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}
```

---

### B7. `toProto(result)` is referenced but never defined [handler/grpc.go:566]

The gRPC handler calls `toProto(result)` to convert `*service.SearchResult` to `*searchv1.SearchResponse`, but this function is never defined in the plan. This is non-trivial because it needs to convert `[]service.Selection` to `[]*searchv1.Selection`, mapping all fields including `json.RawMessage` to `[]byte`, and `[]int` to `[]int32`.

**Fix:** Define the full `toProto` function in the plan. Something like:

```go
func toProto(r *service.SearchResult) *searchv1.SearchResponse {
    resp := &searchv1.SearchResponse{
        Query:      r.Query,
        DurationMs: int32(r.DurationMS),
    }
    for _, s := range r.Selections {
        pages := make([]int32, len(s.Pages))
        for i, p := range s.Pages { pages[i] = int32(p) }
        resp.Selections = append(resp.Selections, &searchv1.Selection{
            Document:   s.Document,
            DocumentId: s.DocumentID,
            NodeIds:    s.NodeIDs,
            Sections:   s.Sections,
            Pages:      pages,
            Text:       s.Text,
            Tables:     s.Tables,
            Images:     s.Images,
        })
    }
    return resp
}
```

---

## Debe corregirse

### D1. Interceptor does not inject role/permissions into context

The HTTP middleware (`pkg/middleware/auth.go` lines 109-110) injects both `WithRole(ctx, claims.Role)` and `WithPermissions(ctx, claims.Permissions)` into the context. The gRPC interceptor only injects tenant info. This means `RequirePermission()` middleware (used by the search HTTP handler at `handler/search.go:33`) would silently fail on gRPC because the role is never in context -- admin users would lose their admin bypass, and all permission checks would deny.

If the gRPC handler calls the same service layer that the HTTP handler calls, and that service layer doesn't do permission checks (it delegates to middleware), then the gRPC handler must either:
(a) Replicate the permission injection in the interceptor, or
(b) Do its own permission check in the gRPC handler before calling the service.

The current search HTTP handler uses `sdamw.RequirePermission("chat.read")` as chi middleware. The gRPC handler has no equivalent check. This means **any authenticated user can call the gRPC search endpoint without `chat.read` permission**.

**Fix:** Either add permission/role injection to the interceptor, or add explicit permission checks in each gRPC handler. Document which approach the plan takes.

---

### D2. No blacklist check in gRPC interceptor

The HTTP `Auth` middleware supports token blacklist checking (`AuthConfig.Blacklist`). The gRPC interceptor has no blacklist support. If a user logs out and their token is blacklisted, the token will be rejected by HTTP endpoints but still accepted by gRPC endpoints until it expires.

**Fix:** Add `*security.TokenBlacklist` parameter to `AuthUnaryInterceptor` and `AuthStreamInterceptor`, or pass it as part of a config struct similar to `AuthConfig`.

---

### D3. No MFA-pending token rejection in gRPC interceptor

The HTTP middleware rejects tokens with `role == "mfa_pending"` (line 90-93 of auth.go). The gRPC interceptor does not check for this. A user mid-MFA flow could use the pending token to call gRPC services.

**Fix:** Add `if claims.Role == "mfa_pending"` check to `extractAndVerifyJWT`.

---

### D4. `Verify` call signature is wrong in interceptor [interceptors.go:366]

The plan's interceptor calls `sdajwt.Verify(token, pub)` with arguments `(string, ed25519.PublicKey)`. The actual signature is:

```go
func Verify(publicKey ed25519.PublicKey, tokenString string) (*Claims, error)
```

The argument order is reversed: `publicKey` is the first parameter, `tokenString` is the second.

**Fix:** Change to `sdajwt.Verify(pub, token)`.

---

### D5. `NewServer` option ordering allows caller to override auth interceptors

```go
return grpc.NewServer(append(defaults, opts...)...)
```

This means any caller-supplied `grpc.ChainUnaryInterceptor` in `opts` will be **appended after** the defaults. Since gRPC chains interceptors in order, the auth interceptor runs first (correct). However, a caller could also pass `grpc.UnaryInterceptor()` (non-chained) which would **replace** the chain entirely, silently removing auth. This is subtle but dangerous.

**Fix:** Document clearly that callers must use `ChainUnaryInterceptor` for additional interceptors, never `UnaryInterceptor`. Or better, reverse the append order so auth is always last (runs first in the chain is actually correct, but document the contract).

---

### D6. Error message in gRPC handler leaks internal details [handler/grpc.go:564]

```go
return nil, status.Errorf(codes.Internal, "search: %v", err)
```

This wraps the internal error directly into the gRPC status message, which is returned to the caller (agent service). While this is internal communication, the error could contain SQL details, file paths, or LLM endpoint URLs. The HTTP handler at `handler/search.go:71` correctly returns a generic `"search failed"` message.

**Fix:** Use a generic message in the gRPC status and log the real error:

```go
slog.Error("search failed", "error", err)
return nil, status.Error(codes.Internal, "search failed")
```

---

### D7. `ForwardJWT` takes raw JWT string -- caller must extract it from incoming context

The `ForwardJWT(ctx, jwt)` function requires the caller to have the raw JWT string. In the HTTP flow, the agent executor gets the JWT from the HTTP `Authorization` header. But when the agent itself is called via gRPC (future), the JWT would be in gRPC metadata, not an HTTP header.

The plan's `GRPCSearchClient.Execute` takes `jwt string` as a parameter, which works for the current HTTP-to-gRPC bridge. But there's no helper to extract the JWT from incoming gRPC metadata for the gRPC-to-gRPC case.

**Fix:** Add a `JWTFromIncomingContext(ctx) (string, error)` helper that extracts the bearer token from incoming gRPC metadata. This is needed when gRPC services chain (e.g., future WS Hub -> Chat via gRPC).

---

### D8. Search service takes a single `*pgxpool.Pool` -- not per-tenant

Looking at `services/search/cmd/main.go`, the search service connects to a single `POSTGRES_TENANT_URL`. The gRPC handler will use this same pool. This means the search gRPC endpoint is already single-tenant by connection. But the plan doesn't address how multi-tenant resolution works when the agent (which is also single-tenant) calls search via gRPC. This is an existing limitation, not introduced by the plan, but the plan should acknowledge it. If search ever needs to resolve tenant pools dynamically, the gRPC interceptor's tenant context injection would need to drive pool resolution.

**Fix:** Add a note to the plan about this -- not a blocker for this plan since the HTTP path has the same limitation, but it should be documented as a known constraint.

---

## Sugerencias

### S1. Scope is correct -- pragmatic and minimal

The plan correctly identifies only 4 HTTP inter-service calls, all from agent. Migrating only the 2 most common ones (search, ingest job status) is the right scope. Keeping notification on NATS and ingest upload on HTTP multipart is pragmatically correct. Not forcing gRPC on auth, platform, or chat (which have no inter-service callers today) follows the bible's principle #1 ("question the requirement before writing code").

### S2. Consider generating `search.proto` during Fase 1, not Fase 3

The plan creates `search.proto` in Fase 3 but could include it in Fase 1's `buf generate` since it's just a proto file. This would make Fase 1 the "all proto work" phase and Fase 3 purely "search Go implementation". Cleaner separation.

### S3. Add gRPC health check service

The interceptor skips `/grpc.health.v1.Health/Check`, but the plan never registers the health check service. Consider using the standard `grpc_health_v1.RegisterHealthServer()` so that container orchestration (Docker health checks, future k8s) can probe gRPC readiness independently of the HTTP `/health` endpoint.

### S4. Consider `buf push` to BSR for remote codegen instead of local `buf generate`

The plan uses remote plugins (`buf.build/protocolbuffers/go`) which require network access at codegen time. If you version the generated code (plan says "generated pero versionado"), consider also storing the `buf.lock` for reproducible builds.

### S5. The `buf.yaml` indentation looks wrong

The plan's `buf.yaml` has `name` at the top level but `modules` also at the top level. In buf v2, `name` should be inside the module definition:

```yaml
version: v2
modules:
  - path: .
    name: buf.build/sda/proto
```

Or `name` is at root level and there's no `modules` block (buf v1 style). Mixing both would be rejected by buf. Verify against buf v2 docs.

### S6. `search.proto` design -- missing `tenant_id` or `tenant_slug` field

The current `SearchRequest` has no tenant identification field. The plan relies on the JWT interceptor to inject tenant context. This is correct for the auth flow, but it means the proto message alone doesn't carry enough information for observability/debugging. Consider adding a `string tenant_slug = 4` field even if it's set server-side from JWT context. The existing protos (`ingest.proto`, `chat.proto`) include `user_id` fields "set by auth middleware" -- search.proto could follow the same pattern for consistency.

### S7. Phase estimates are realistic

2-3 days total is achievable for someone familiar with the codebase. The heaviest phases are 2 (pkg/grpc) and 3 (search gRPC server) at ~3-4h each. The only risk is debugging buf codegen issues if the proto files have subtle incompatibilities.

### S8. Consider graceful shutdown ordering

The plan adds `grpcSrv.GracefulStop()` to the shutdown sequence but doesn't specify the order relative to `srv.Shutdown()`. The correct order is: (1) stop accepting new gRPC connections, (2) stop accepting new HTTP connections, (3) drain in-flight requests. Since `GracefulStop()` blocks, it should be called before `srv.Shutdown()`, or both should be done concurrently.

### S9. `CollectionId: &p.CollectionID` sends empty string pointer when no collection specified

In `GRPCSearchClient.Execute` (line 672), the code always passes `&p.CollectionID` even when it's empty. The proto uses `optional string`, so passing a pointer to an empty string is semantically different from `nil` (no value). The search service treats `""` as "no filter" but this semantic coupling is fragile.

**Fix:** Only set `CollectionId` if non-empty:

```go
req := &searchv1.SearchRequest{Query: p.Query, MaxNodes: int32(p.MaxNodes)}
if p.CollectionID != "" {
    req.CollectionId = &p.CollectionID
}
```

### S10. The plan should address what happens to `rag.proto`

The plan mentions deprecating `rag.proto` in Fase 7 (add a comment). But it also generates code from it (`gen/go/rag/v1/*.pb.go`). Since the rag service is deprecated and replaced by agent+search, consider NOT generating code from `rag.proto` at all. Or at minimum, don't generate the `_grpc.pb.go` server stub to avoid confusion. You can exclude it in `buf.gen.yaml` using a module or path exclude.

### S11. `opts...` pattern in `Dial` and `NewServer` should validate no duplicates

Both `Dial` and `NewServer` use `append(defaults, opts...)`. If a caller passes `grpc.WithTransportCredentials(...)`, they get two transport credentials options, and gRPC silently uses the last one. This could accidentally enable TLS or something unexpected. Consider logging or documenting this.

---

## Lo que esta bien

1. **The scope decision is excellent.** Only migrating calls that actually exist (agent -> search, agent -> ingest job status), keeping everything else as-is, and explicitly listing what's out of scope with reasons. This follows bible principle #1 perfectly.

2. **The "Que NO resuelve" section is clear.** No service mesh, no mTLS, no gRPC-Web -- all correct exclusions at this stage.

3. **JWT forwarding via metadata is the right approach.** Using the same JWT from the original user request means no service-to-service token machinery, no token exchange protocol, and the callee verifies the same claims as if the user called directly. Simple and secure.

4. **Fallback to HTTP if gRPC unavailable** is a solid operational decision. It means gRPC can be rolled out gradually without hard cutover.

5. **gRPC ports hidden from Traefik and host in prod** is correct. The `expose:` vs `ports:` distinction in docker-compose.prod.yml is the right mechanism.

6. **The proto file inventory is accurate.** All 6 existing protos verified -- line counts, RPC counts, and `go_package` values match reality.

7. **The `wrappedStream` pattern for stream interceptors** is the canonical gRPC approach. Correctly implemented.

8. **Agent executor HTTP calls accurately described.** The 4 calls listed match exactly what's in `services/agent/cmd/main.go` lines 68-85.

9. **The dependency chain is correctly ordered.** Fase 1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 7 has no hidden circular dependencies. Each phase's prerequisites are accurately stated.

10. **Not generating code for protos that have no consumers** is pragmatic. Auth, chat, platform, notification proto code will be generated but no gRPC servers implemented -- this is fine as infrastructure for future use.
