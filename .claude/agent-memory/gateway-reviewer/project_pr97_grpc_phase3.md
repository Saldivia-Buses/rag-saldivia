---
name: PR #97 gRPC Phase 3
description: Plan 09 Phase 3 -- first gRPC server (search) + interceptor fixes. CAMBIOS REQUERIDOS. Blocker: no permission check on gRPC Query. Must-fix: ValidToken test hollow, HTTP middleware missing UserID/Email context injection.
type: project
---

PR #97 reviewed 2026-04-05. Result: CAMBIOS REQUERIDOS.

**Blocker:**
1. `grpc.go:25` -- gRPC Query RPC has no RequirePermission("chat.read") check. HTTP handler has it (search.go:33). Any valid JWT can call the RPC regardless of permissions. Add sdamw.RoleFromContext + sdamw.PermissionsFromContext check before delegating to service.

**Must-fix:**
2. `grpc_test.go:21` ValidToken test has no value assertions -- does not verify UserID/Email/Role/TenantID were actually injected. Test would pass if injections were deleted.
3. `pkg/middleware/auth.go` HTTP middleware still does not call WithUserID/WithUserEmail into context. gRPC interceptor does. sdamw.UserIDFromContext() silently returns "" on HTTP paths. No current handler uses it yet, but the helper is now live.

**Suggestions:** Dead code lines 91-92 in WrongKey test; ErrServerStopped not filtered in grpc serve goroutine; slog.Error in grpc.go missing user_id field.

**What's good:** toProto types correct, SearchDocuments signature matches, dual listener + shutdown order correct, UnimplementedSearchServiceServer embed present, interceptor now has full parity (UserID+Email+Role+Permissions+Tenant).

**Why:** Permission check gap is a blocker because gRPC callers bypass RBAC. HTTP parity gap matters because any future handler using UserIDFromContext will silently break on HTTP paths.

**How to apply:** When future gRPC handler PRs arrive, always check for explicit permission guard matching the HTTP route. When auth.go is touched, verify UserID/Email are injected into context.
