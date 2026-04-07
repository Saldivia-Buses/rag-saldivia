# Gateway Review — PR #97 Plan 09 Phase 3 (gRPC server + interceptor fixes)

**Fecha:** 2026-04-05
**Resultado:** CAMBIOS REQUERIDOS

## Bloqueantes

**1. `grpc.go:25` — gRPC Query RPC has no permission check (HTTP parity gap)**

The HTTP handler guards `POST /v1/search/query` with `RequirePermission("chat.read")`
(`search.go:33`). The gRPC `Query` RPC has no equivalent check — any caller with
a valid JWT (any role, any permission set) can invoke it over gRPC.

Fix: add an explicit check at the top of `Query` before delegating to the service:

```go
import sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"

func (h *GRPCHandler) Query(ctx context.Context, req *searchv1.SearchRequest) (*searchv1.SearchResponse, error) {
    if req.Query == "" {
        return nil, status.Error(codes.InvalidArgument, "query is required")
    }
    role  := sdamw.RoleFromContext(ctx)
    perms := sdamw.PermissionsFromContext(ctx)
    if role != "admin" && !containsPerm(perms, "chat.read") {
        return nil, status.Error(codes.PermissionDenied, "insufficient permissions")
    }
    // rest unchanged
}
```

---

## Debe corregirse

**2. `grpc_test.go:21` — ValidToken test does not assert injected values**

The entire goal of this PR's interceptor fix is that UserID and Email are now
injected into context. `TestExtractAndVerifyJWT_ValidToken` only checks that no
error is returned and context is non-nil. It would pass even if the
`WithUserID`/`WithUserEmail` lines were deleted.

Add value assertions for every injected field:

```go
if sdamw.UserIDFromContext(newCtx) != "u-1" {
    t.Errorf("UserID: got %q, want 'u-1'", sdamw.UserIDFromContext(newCtx))
}
if sdamw.UserEmailFromContext(newCtx) != "test@sda.app" {
    t.Errorf("Email: got %q, want 'test@sda.app'", sdamw.UserEmailFromContext(newCtx))
}
if sdamw.RoleFromContext(newCtx) != "admin" {
    t.Errorf("Role: got %q, want 'admin'", sdamw.RoleFromContext(newCtx))
}
ti, _ := tenant.FromContext(newCtx)
if ti.ID != "t-1" {
    t.Errorf("TenantID: got %q, want 't-1'", ti.ID)
}
```

**3. `pkg/middleware/auth.go` — HTTP middleware does not inject UserID/Email into context**

The gRPC interceptor calls `WithUserID(ctx, claims.UserID)` and
`WithUserEmail(ctx, claims.Email)` (interceptors.go:109-110). The HTTP
middleware does NOT — it sets `X-User-ID`/`X-User-Email` headers but skips
the context helpers. The two paths are now asymmetric.

No HTTP handler currently reads `UserIDFromContext(ctx)` yet, so there's no
immediate breakage. But with `WithUserID`/`WithUserEmail` now exported from
`pkg/middleware`, future handlers will use them and silently get empty string
on the HTTP path.

Fix: in `pkg/middleware/auth.go`, alongside the existing `WithRole`/`WithPermissions`
calls, add:

```go
ctx = WithUserID(ctx, claims.UserID)
ctx = WithUserEmail(ctx, claims.Email)
```

---

## Sugerencias

- `grpc_test.go:91-92` — The `cfg` assignment is dead code. Lines 91-92 create
  and immediately discard a `DefaultConfig`. The token is signed on line 93 using
  a separate inline config call. Remove lines 91-92.

- `main.go:109` — When `grpcSrv.GracefulStop()` is called, `Serve()` returns
  `grpc.ErrServerStopped`. The goroutine logs it as an error even though it's
  expected. Filter it:
  `if err := grpcSrv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped)`

- `grpc.go:37` — `slog.Error` on search failure logs only `error`. The HTTP
  handler includes `request_id` and `tenant_id`. Consider adding
  `sdamw.UserIDFromContext(ctx)` to the log entry for traceability.

---

## Lo que está bien

- `toProto` conversion is type-correct. `service.Selection.Tables` and `.Images`
  are `json.RawMessage` (underlying `[]byte`), matching `proto.Selection.Tables`
  and `.Images` (`[]byte`). Direct assignment compiles without conversion.

- `SearchDocuments` call signature matches exactly: `(ctx, query string,
  collectionID string, maxNodes int)` — the `int(req.MaxNodes)` cast from `int32`
  is correct, and `maxNodes <= 0` defaults to 5 so an unset proto field is handled.

- Dual listener pattern is correct. Both servers start in goroutines sharing one
  `signal.NotifyContext`. On shutdown, `GracefulStop()` drains gRPC before
  `srv.Shutdown()` drains HTTP. Order is right.

- `UnimplementedSearchServiceServer` embed ensures forward compatibility when
  new RPCs are added to the proto without recompiling all servers.

- `contextKey` type in `rbac.go` uses the unexported string type — no collision
  with `auth.go`'s context keys possible within the package.

- The `rbac.go` additions (`WithUserID`, `WithUserEmail`, `UserIDFromContext`,
  `UserEmailFromContext`) close the PR #96 follow-up gap. The interceptor now
  injects all five identity fields: tenant, role, permissions, user ID, email.

- All 10 tests use real Ed25519 keys generated inline — no mocks, no hardcoded
  key material. MFA-pending rejection, wrong-key rejection, blacklist path,
  ForwardJWT round-trip, and JWTFromIncomingContext are all covered.
