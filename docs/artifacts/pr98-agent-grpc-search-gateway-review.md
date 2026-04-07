# Gateway Review — PR #98: Agent gRPC Client for Search (Plan 09 Phase 4)

**Fecha:** 2026-04-05
**Resultado:** APROBADO

## Bloqueantes

None.

## Debe corregirse

**1. `grpc_search.go:47` — JSON unmarshal error leaks raw Go error text to LLM context**

```go
return &Result{Status: "error", Error: "invalid search params: " + err.Error()}, nil
```

`err.Error()` from `json.Unmarshal` includes field paths and type info. This is not a security issue (it never leaves the agent service), but it gives the LLM confusing internal error strings. Fix: generic message.

```go
return &Result{Status: "error", Error: "invalid search params"}, nil
```

**2. `grpc_search.go:67` — `json.Marshal(resp)` error silently dropped**

```go
data, _ := json.Marshal(resp)
```

`searchv1.SearchResponse` is a protobuf message. `encoding/json` can fail on proto messages with `oneof` or `bytes` fields (and silently returns `null`). The LLM would receive `{"status":"success","data":null}` and hallucinate. Use `google.golang.org/protobuf/encoding/protojson` instead, which is the correct serializer for proto messages.

```go
data, err := protojson.Marshal(resp)
if err != nil {
    return &Result{Status: "error", Error: "search failed"}, nil
}
return &Result{Status: "success", Data: json.RawMessage(data)}, nil
```

**3. `executor.go:103` — `ExecuteConfirmed` bypasses gRPC path for `search_documents`**

```go
func (e *Executor) ExecuteConfirmed(...) (*Result, error) {
    ...
    return e.executeHTTP(ctx, jwt, def, params)
}
```

`search_documents` has `RequiresConfirmation: false` in the current toolDefs (main.go:74), so this path is unreachable today. But `ExecuteConfirmed` always goes to HTTP even if gRPC is wired. If `search_documents` ever gains `RequiresConfirmation: true`, the fallback silently goes HTTP. Add the same gRPC check:

```go
if toolName == "search_documents" && e.grpcSearch != nil {
    return e.grpcSearch.Execute(ctx, jwt, params)
}
```

## Sugerencias

- `grpc_search.go:63`: `slog.Warn` is correct for degraded-path logging, but add `"tool", "search_documents"` field for easier log correlation — the agent may call many tools per turn and you want to attribute failures quickly.
- `main.go:106-116`: The non-fatal gRPC init is well-structured. Consider logging the fallback path at INFO level when `SEARCH_GRPC_URL` is empty (not just when it fails) so it's obvious in startup logs which transport is active.

## Lo que está bien

- **JWT forwarding is correct.** `sdagrpc.ForwardJWT(ctx, jwt)` appends `authorization: Bearer <jwt>` to outgoing metadata. The search service interceptor (`extractAndVerifyJWT`) reads `md.Get("authorization")`, strips the `Bearer ` prefix, and verifies with Ed25519. Full round-trip works.
- **Fallback is transparent.** `Execute` checks `e.grpcSearch != nil` before routing — if the client never wired (empty env) or init failed (non-fatal path in main.go), it falls through to `executeHTTP` with identical parameters. No behavior change from caller perspective.
- **`CollectionId` optional handling is correct.** `if p.CollectionID != ""` sets the proto pointer correctly; zero-value means the server treats it as unset. Matches how the gRPC handler reads it (`if req.CollectionId != nil`).
- **`MaxNodes` zero-value is safe.** Proto default is 0, which `service.SearchDocuments` handles as "use service default" (no explicit validation needed on client side).
- **Non-fatal gRPC init is correct pattern.** Search via gRPC is an optimization, not a requirement. Warn + HTTP fallback is the right behavior for a missing/unreachable gRPC endpoint at startup.
- **Permission check is on the server, not the client.** `grpc.go:29-40` has the `chat.read` guard that was missing in PR #97. This PR does not need to re-implement it on the client.
- **`defer grpcClient.Close()` is correct** — connection is closed on graceful shutdown after `srv.Shutdown` completes.
- **`sdagrpc.Dial` is lazy** — does not fail if search is unreachable at agent startup. This is documented and correct.
