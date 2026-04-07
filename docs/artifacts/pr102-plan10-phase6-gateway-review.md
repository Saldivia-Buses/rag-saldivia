# Gateway Review — PR #102 Plan 10 Phase 6: Chat gRPC Server + WS Hub Mutations

**Fecha:** 2026-04-05
**Resultado:** CAMBIOS REQUERIDOS

---

## Bloqueantes

### B1 — gRPC error message leaks to WebSocket client
**File:** `services/ws/internal/hub/mutations.go:56`

`Handle()` sends `err.Error()` verbatim to the connected client:

```go
client.SendMessage(Message{
    Type:  Error,
    ID:    msg.ID,
    Error: err.Error(),
})
```

When the Chat gRPC server returns a gRPC status error (e.g. `codes.Internal: "create session failed"`), `err.Error()` produces `"rpc error: code = Internal desc = create session failed"` — the full gRPC status string. When Chat is unreachable the transport error is worse (`connection refused`, or the full dial/TLS error). Neither is a safe message to give a browser client.

The `unknownActionError` at line 128 also leaks the raw action string, which is client-supplied input.

**Fix:** map gRPC codes to clean client messages in `dispatch`, and sanitize `unknownActionError`:

```go
import "google.golang.org/grpc/status"
import "google.golang.org/grpc/codes"

func toClientError(err error) string {
    st, ok := status.FromError(err)
    if !ok {
        return "internal error"
    }
    switch st.Code() {
    case codes.NotFound:
        return "not found"
    case codes.PermissionDenied, codes.Unauthenticated:
        return "unauthorized"
    case codes.InvalidArgument:
        return "invalid request"
    default:
        return "internal error"
    }
}
```

Use `toClientError(err)` instead of `err.Error()` in `Handle()`. For `unknownActionError`, the action value is already validated by the switch default, but the string still exposes client-controlled text — use a fixed message: `"unknown action"`.

---

### B2 — JWT stored in Client is never refreshed; mutations silently fail or leak auth errors after 15 min

**Files:** `services/ws/internal/hub/client.go:35`, `services/ws/internal/hub/mutations.go:70`

The Client stores the raw access token on connection (`JWT string`). Access tokens expire in 15 minutes (per system spec). After expiry, `dispatch()` calls `sdagrpc.ForwardJWT(context.Background(), client.JWT)` with the stale token. The Chat gRPC interceptor will return `codes.Unauthenticated`. The client receives an error (leaking the gRPC status per B1 above), with no indication that re-authentication is needed. The WS connection itself remains open.

This was flagged as B3 in the Plan 10 spec review and is still unresolved.

There are two acceptable strategies:

**Option A — Client sends refresh token via mutation, hub updates `Client.JWT`.**  
Add a `refresh_token` action to `handleMessage`. Client sends `{type:"mutation", action:"refresh_token", data:{token:"..."}}`. Hub verifies the new token with `sdajwt.Verify`, updates `client.JWT`, responds with `{type:"event", ...}`. No gRPC call needed — all logic stays in WS Hub.

**Option B — Hub detects expired-token gRPC errors and sends a structured `token_expired` event.**  
On `codes.Unauthenticated` from Chat, send `{type:"error", code:"token_expired"}` so the frontend knows to re-auth. This is the minimum viable fix: no new mutation needed, the client handles the reconnect. Still requires B1 to be fixed first so the error code reaches the client cleanly.

Option B is simpler and can be implemented now. Option A is complete but more surface. Pick one and document it — the current behavior (silent auth failure buried in a generic error string) is not acceptable.

---

## Debe corregirse

### M1 — `grpc.go`: UserID fallback to `req.UserId` is a privilege escalation path

**File:** `services/chat/internal/handler/grpc.go:27-30` (and same pattern in all 7 RPCs)

```go
userID := sdamw.UserIDFromContext(ctx)
if userID == "" {
    userID = req.UserId
}
```

The gRPC interceptor in `pkg/grpc/interceptors.go` always injects `UserID` via `sdamw.WithUserID(ctx, claims.UserID)` when auth passes. If `UserIDFromContext` returns `""` it means the interceptor did NOT run — either the server was started without the interceptor, or the interceptor config is wrong.

The fallback `req.UserId` should never be reached on a correctly-configured server, but if it is reached, it allows the caller to supply any arbitrary user ID. This is the same class of bug as HTTP header spoofing, just one layer down.

The correct fix is to remove the fallback and fail hard:

```go
userID := sdamw.UserIDFromContext(ctx)
if userID == "" {
    return nil, status.Error(codes.Unauthenticated, "missing user identity")
}
```

This will surface misconfiguration at call time rather than silently accepting spoofed identity. The existing integration tests should cover this path.

### M2 — `grpc.go`: `AddMessage` does not validate `role`

**File:** `services/chat/internal/handler/grpc.go:115`

```go
m, err := h.svc.AddMessage(ctx, req.SessionId, userID, req.Role, req.Content, nil, req.Sources, req.Metadata)
```

`req.Role` is passed directly. The proto comment says `"user", "assistant", or "system"` are the valid values. The service layer does not validate role — it goes directly to `repo.CreateMessage`. A WS client sending `{action:"send_message", data:{role:"system"}}` can inject a `system` role message.

The HTTP handler presumably validates this somewhere. Check and replicate the validation here:

```go
switch req.Role {
case "user", "assistant", "system":
    // ok
default:
    return nil, status.Error(codes.InvalidArgument, "invalid role")
}
```

### M3 — `grpc.go`: `ListMessages` does a session ownership check but passes the wrong limit

**File:** `services/chat/internal/handler/grpc.go:137`

```go
messages, err := h.svc.GetMessages(ctx, req.SessionId, 100)
```

The proto `ListMessagesRequest` has a `limit` field (field 3 in the generated code). It is silently ignored and hardcoded to 100. This diverges from HTTP behavior and will surprise callers. Use `req.Limit` with a safe cap:

```go
limit := req.Limit
if limit <= 0 || limit > 200 {
    limit = 100
}
messages, err := h.svc.GetMessages(ctx, req.SessionId, limit)
```

### M4 — `mutations.go`: `context.Background()` ignores request deadline

**File:** `services/ws/internal/hub/mutations.go:70`

```go
ctx := sdagrpc.ForwardJWT(context.Background(), client.JWT)
```

`context.Background()` has no deadline. If Chat is slow or hung, this goroutine blocks forever with no timeout. The WS hub has no way to bound the number of in-flight mutation goroutines, so a slow Chat service can exhaust goroutines under load.

Use a bounded context:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
ctx = sdagrpc.ForwardJWT(ctx, client.JWT)
```

---

## Sugerencias

- `chat/cmd/main.go:117` — the gRPC goroutine calls `os.Exit(1)` on `net.Listen` failure. At that point the HTTP server and NATS connection are already initialized. Consider returning the error and doing a clean `cancel()` + drain instead, so the defer chains run.

- `hub.go` has no `Mutations` concurrency concern: `h.Mutations` is set once before `h.Run()` starts (in `ws/cmd/main.go:74-77`) and never mutated again. The nil check at line 177 is safe as-is. No issue.

- The goroutine dispatch in `mutations.Handle()` is safe with respect to the `client` pointer. `client.SendMessage` is concurrency-safe (uses `TrySend` + `closed` atomic). No race condition. The `client.JWT` and `client.UserID` fields are set once at construction and never written again. Safe.

- `grpc.go`: `errors.Is` should be preferred over `==` for sentinel errors to be future-proof: `if errors.Is(err, service.ErrSessionNotFound)`. Currently using `==` at lines 53, 85, 101.

---

## Lo que está bien

- `chat/cmd/main.go`: dual-listener pattern is clean. gRPC and HTTP share the same service layer correctly. `grpcSrv.GracefulStop()` is called before `srv.Shutdown()` — correct ordering.
- `mutations.go`: `NewMutationHandler` returns nil gracefully when `CHAT_GRPC_URL` is unset. `hub.go:177` checks nil before dispatch. WS Hub degrades gracefully with `"mutations not available"`.
- `mutations.go`: uses `protojson.Marshal` for proto responses — correct (avoids the silent-null problem with `encoding/json` on proto messages, as flagged in PR #98 review).
- `client.go`: `JWT` field is a raw string, not a pointer, set once in constructor and never written again. No mutex needed for reads. Correct.
- `ws/cmd/main.go`: `CHAT_GRPC_URL` defaults to `""` and the nil-return path is handled. The mutation handler is wired after the hub is created but before `h.Run()` is started — no race on `h.Mutations`.
- `grpc.go`: `sessionToProto` and `messageToProto` correctly map all domain fields. `timestamppb.New()` is the right API for time conversion.
- Auth interceptor in `pkg/grpc/interceptors.go` strips MFA-pending tokens — parity with HTTP middleware is maintained.
