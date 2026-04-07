# Gateway Review -- PR #78 Service Wiring Expansion

**Fecha:** 2026-04-05
**Branch:** feat/service-wiring-expansion
**Resultado:** CAMBIOS REQUERIDOS

## Resumen

PR addresses P0/P1 findings from a service analysis audit across three areas:
1. Agent: new `TracePublisher` for NATS trace events, wired into the query loop
2. Chat: guardrails on user messages in `AddMessage`
3. Feedback: Dockerfile fixed to match project pattern

---

## Bloqueantes

### B1. TracePublisher hardcodes tenant slug as `"default"` -- tenant isolation broken

**Files:**
- `services/agent/internal/service/agent.go:85`
- `services/agent/internal/service/agent.go:260`

All trace events are published with `"default"` as the tenant slug:

```go
traceID := a.tracePublisher.TraceStart("default", "", "", userMessage)
// ...
a.tracePublisher.TraceEnd("default", traceID, status, ...)
```

The comment on line 84 acknowledges this: "tenant extracted from JWT would be ideal, using 'default' for now".

**Impact:** Every trace from every tenant writes to `tenant.default.traces.*` subjects. The Traces Service subscriber will store them all with `tenant_id = "default"` in the platform DB. This means:
- Cross-tenant data leak: all tenants' queries are visible in one bucket
- `ListTraces(ctx, tenantID, ...)` will return nothing for any real tenant
- Cost attribution is impossible (all costs assigned to "default")

**Fix:** The `Query` method already receives the request context (which has `tenant.Info` injected by `pkg/middleware.Auth`). Extract it:

```go
func (a *Agent) Query(ctx context.Context, jwt, userMessage string, history []llm.Message) (*QueryResult, error) {
    start := time.Now()

    ti, err := tenant.FromContext(ctx)
    if err != nil {
        return nil, fmt.Errorf("tenant context required: %w", err)
    }
    slug := ti.Slug

    traceID := a.tracePublisher.TraceStart(slug, "", "", userMessage)
    // ... and update publishTraceEnd to use slug instead of "default"
```

Also pass `sessionID` and `userID` from context headers instead of empty strings -- the Traces Service schema expects these for meaningful trace records.

### B2. TracePublisher has no NATS subject injection validation

**File:** `services/agent/internal/service/traces.go:35,55,71`

The `TraceStart`, `TraceEnd`, and `TraceEvent` methods build NATS subjects via `fmt.Sprintf("tenant.%s.traces.start", tenantSlug)` without any validation of `tenantSlug`.

The shared `pkg/nats/publisher.go` has `isValidSubjectToken()` that rejects empty strings and strings containing `.*> \t\r\n`. The `TracePublisher` bypasses this entirely.

**Impact:** If `tenantSlug` contains NATS wildcard characters (`.`, `*`, `>`), a crafted slug could subscribe to or publish on arbitrary NATS subjects. While currently hardcoded to `"default"`, once B1 is fixed and real slugs flow through, this becomes exploitable if a slug is ever malformed.

**Fix:** Either:
1. Use `pkg/nats/publisher.go`'s `Publisher` or export `isValidSubjectToken` and call it, or
2. Copy the validation inline:

```go
func (p *TracePublisher) publish(subject string, evt any) {
    // ... existing code
}

func (p *TracePublisher) validSlug(slug string) bool {
    return slug != "" && !strings.ContainsAny(slug, ".*> \t\r\n")
}
```

And validate in each public method before publishing.

### B3. `publishTraceEnd` not called on `pending_confirmation` return path

**File:** `services/agent/internal/service/agent.go:198-215`

When a tool returns `pending_confirmation`, the method returns early without calling `publishTraceEnd`:

```go
} else if result.Status == "pending_confirmation" {
    tcLog.Status = "pending_confirmation"
    allToolCalls = append(allToolCalls, tcLog)
    return &QueryResult{...}, nil  // <-- no TraceEnd published
}
```

**Impact:** The trace stays in `status = 'running'` forever in the Traces Service DB. If the user never confirms, the trace is orphaned. The `GetTenantCost` query filters by `status = 'completed'`, so this won't inflate costs, but `ListTraces` will accumulate stale "running" traces.

**Fix:** Publish a trace end with `status = "pending_confirmation"` before returning:

```go
} else if result.Status == "pending_confirmation" {
    tcLog.Status = "pending_confirmation"
    allToolCalls = append(allToolCalls, tcLog)
    r := &QueryResult{...}
    return a.publishTraceEnd(traceID, r, "pending_confirmation"), nil
}
```

---

## Debe corregirse

### C1. `TraceEvent` method is dead code -- individual step tracing not wired

**File:** `services/agent/internal/service/traces.go:58-72`

`TraceEvent` is defined but never called from `agent.go`. The agent loop does LLM calls and tool calls but never publishes per-step trace events.

**Impact:** The Traces Service has a `trace_events` table expecting per-step records (llm_call, tool_call, error). Without these events, the trace detail view (`GetTraceDetail`) returns an empty events list -- the feature is inert.

**Fix:** Publish `TraceEvent` entries inside the agent loop:
- After each LLM call (`resp` from `a.llmAdapter.Chat`)
- After each tool execution (with tool name, status, duration)

Example for tool calls (inside the for loop, after tool execution):

```go
a.tracePublisher.TraceEvent(slug, traceID, "tool_call", len(allToolCalls), tcDuration, map[string]any{
    "tool": tc.Function.Name,
    "status": tcLog.Status,
})
```

### C2. `TraceStart` passes empty `sessionID` and `userID`

**File:** `services/agent/internal/service/agent.go:85`

```go
traceID := a.tracePublisher.TraceStart("default", "", "", userMessage)
```

Both `sessionID` and `userID` are empty strings. The Traces Service stores these in the DB and they're part of the `ListTraces` / `GetTraceDetail` response.

**Fix:** Extract from request context or handler:
- `userID` is available via `r.Header.Get("X-User-ID")` -- pass it through to the service method
- `sessionID` may not apply to agent queries (no chat session), but if one exists, pass it

The `Query` signature should be extended or the context should carry these values.

### C3. Chat guardrails config is hardcoded in the handler, not configurable

**File:** `services/chat/internal/handler/chat.go:215-218`

```go
sanitized, err := guardrails.ValidateInput(r.Context(), req.Content, guardrails.InputConfig{
    MaxLength:     50000,
    BlockPatterns: []string{"ignora tus instrucciones", "ignore your instructions"},
}, nil)
```

The `InputConfig` is hardcoded inline with:
- Fixed max length of 50000
- Only 2 block patterns (same ones the agent uses)
- No LLM classifier (nil)

**Impact:** Block patterns cannot be updated without redeploying the service. Different tenants might need different patterns. The patterns are trivially bypassable (e.g., "ignora tus instrucciones" vs "ignorar tus instrucciones" or "ignorA TUS instruccIones").

**Fix:** Accept the config from the `Chat` struct (injected at startup from env/config), matching the pattern used in the agent service:

```go
type Chat struct {
    chatSvc          ChatService
    guardrailsConfig guardrails.InputConfig
}
```

And use `h.guardrailsConfig` in the handler instead of the inline literal.

### C4. Chat guardrails only apply to `role == "user"` but `system` role is still allowed

**File:** `services/chat/internal/handler/chat.go:190,208-209,214`

```go
var validRoles = map[string]bool{"user": true, "assistant": true, "system": true}
// ...
if req.Role == "user" {
    // guardrails validation
}
```

The `system` role is in `validRoles` and passes through WITHOUT guardrails. A client can submit `role: "system"` messages that go directly into the chat session unfiltered.

**Impact:** An attacker can store arbitrary "system" messages in a chat session. When these messages are later sent as context to the LLM (via the RAG or agent services), they act as injected system prompts.

**Fix:** Either remove `system` from `validRoles` (API clients should not submit system messages) or apply guardrails to `system` role messages too. Removing is safer:

```go
var validRoles = map[string]bool{"user": true, "assistant": true}
```

---

## Sugerencias

### S1. Consider using `pkg/nats.Publisher` instead of a custom TracePublisher

The `pkg/nats/publisher.go` already handles slug validation, JSON marshaling, and subject formatting. The `TracePublisher` reimplements all of this. Consider either:
- Using `Broadcast` with a structured channel like `traces.start`
- Or extending `Publisher` with a `PublishRaw(slug, subject, data)` method

This avoids maintaining duplicate NATS publishing logic.

### S2. Add `trace_id` to agent query response

The `QueryResult` struct does not include `traceID`. The frontend or calling service has no way to correlate an agent query with its trace. Consider adding it to the response.

### S3. Chat guardrails should run BEFORE ownership check

**File:** `services/chat/internal/handler/chat.go:214-234`

Currently the order is: parse body -> validate roles -> guardrails -> ownership check -> AddMessage. This is fine -- guardrails running before the DB call is correct. Just noting the sequence is good.

### S4. Feedback Dockerfile looks correct

The Dockerfile follows the exact same pattern as `services/auth/Dockerfile`:
- go.work + pkg + service-specific dir copied
- distroless base image
- nonroot user
- Correct port 8008 exposed

No issues found.

---

## Lo que esta bien

1. **TracePublisher nil safety**: `NewTracePublisher` accepts nil conn, all methods check `p.nc == nil` before publishing -- graceful degradation when NATS is unavailable.

2. **NATS connection in agent main.go**: Uses `nats.MaxReconnects(-1)` + `ReconnectWait(2s)` for resilient reconnection. Connection failure is a warning, not fatal -- the service still works without tracing.

3. **Chat guardrails placement**: Applied at the handler level before hitting the service/DB layer. Uses `guardrails.ValidateInput` correctly with context. Blocked messages return 400 with a generic error (no leak of what pattern matched).

4. **Agent main.go middleware chain**: Correct order (RequestID -> RealIP -> Recoverer -> Timeout -> SecureHeaders -> Auth group). Health endpoint excluded from auth.

5. **Feedback Dockerfile**: Clean, matches project convention exactly. Uses distroless, nonroot, correct port.

6. **Agent NATS cleanup**: Uses `nc.Drain()` on shutdown, which is the correct NATS cleanup method (flushes pending publishes before closing).

7. **Chat MaxBytesReader**: Present on AddMessage (`1<<20`), preventing oversized payloads.

---

## Resumen de hallazgos

| ID | Severidad | Archivo | Descripcion |
|----|-----------|---------|-------------|
| B1 | BLOQUEANTE | agent/service/agent.go:85,260 | Tenant slug hardcoded "default" -- tenant isolation broken |
| B2 | BLOQUEANTE | agent/service/traces.go:35,55,71 | No NATS subject injection validation |
| B3 | BLOQUEANTE | agent/service/agent.go:198-215 | TraceEnd not published on pending_confirmation |
| C1 | DEBE CORREGIR | agent/service/traces.go:58-72 | TraceEvent is dead code, per-step tracing not wired |
| C2 | DEBE CORREGIR | agent/service/agent.go:85 | Empty sessionID and userID in TraceStart |
| C3 | DEBE CORREGIR | chat/handler/chat.go:215-218 | Guardrails config hardcoded, not configurable |
| C4 | DEBE CORREGIR | chat/handler/chat.go:190,214 | System role allowed without guardrails |
