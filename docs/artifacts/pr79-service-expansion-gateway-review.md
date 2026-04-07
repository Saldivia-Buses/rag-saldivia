# Gateway Review -- PR #79 Service Expansion P1/P2

**Fecha:** 2026-04-05
**Branch:** feat/service-expansion-p1p2
**Resultado:** CAMBIOS REQUERIDOS

---

## Bloqueantes

### B1. ExtractorConsumer: tenant isolation broken -- single pool for all tenants
**File:** `/home/enzo/rag-saldivia/services/ingest/internal/service/extractor_consumer.go:28-35`

The consumer subscribes to `tenant.*.extractor.result.>` (all tenants), but is constructed with a single `*pgxpool.Pool`. When messages from tenant B arrive, they get written to tenant A's database. This is a hard tenant isolation violation.

**Fix:** The consumer needs access to `pkg/tenant.Resolver` to resolve the correct pool per tenant. Extract the tenant slug from the NATS subject (like the ingest worker does in `worker.go:130`) and call `resolver.Pool(slug)` per message. Constructor should take `*tenant.Resolver` instead of `*pgxpool.Pool`.

```go
func NewExtractorConsumer(nc *nats.Conn, resolver *tenant.Resolver, treeGen *tree.Generator) *ExtractorConsumer {
```

In `handleResult`, extract tenant from the NATS subject and resolve the pool:
```go
slug := tenantFromSubject(msg.Subject())
pool, err := c.resolver.Pool(ctx, slug)
if err != nil { msg.Nak(); return }
repo := repository.New(pool)
```

### B2. ExtractorConsumer: DB errors silently swallowed, document marked "ready" on partial failure
**File:** `/home/enzo/rag-saldivia/services/ingest/internal/service/extractor_consumer.go:110-170`

Lines 111-136: All `repo.UpdateDocumentStatus`, `repo.UpdateDocumentPages`, `repo.InsertDocumentPage` return errors that are completely ignored. If any INSERT fails (e.g., constraint violation, pool timeout), the code continues to the next step and eventually marks the document as "ready" and Ack's the message.

Worst case: tree generation succeeds but zero pages are actually stored -- document appears "ready" but search returns nothing.

**Fix:** Check every error. On failure, log the error, Nak the message (for retry), and return:

```go
if err := c.repo.UpdateDocumentStatus(ctx, ...); err != nil {
    slog.Error("update document status", "doc_id", result.DocumentID, "error", err)
    msg.Nak()
    return
}
```

If page insertion partially fails, mark document as "error" (not "ready"), Nak, and return.

### B3. Platform publishLifecycleEvent: NATS subject injection via event type
**File:** `/home/enzo/rag-saldivia/services/platform/internal/service/platform.go:52-62`

The call `p.publisher.Notify(tenantSlug, map[string]any{"type": "platform." + eventType, ...})` passes event type `"platform.tenant.created"` which contains dots. `pkg/nats/publisher.go:57` uses `parsed.Type` directly in the NATS subject WITHOUT validation:

```
tenant.{slug}.notify.platform.tenant.created
```

This creates 3 extra subject levels instead of 1. Any subscriber on `tenant.*.notify.>` will still receive it, but it breaks the expected subject hierarchy and consumers that filter on `tenant.*.notify.*` (single token) will miss these events entirely.

This is a **known recurring issue** (see PR #52 review). The fix belongs in `pkg/nats/publisher.go` -- `Notify` should validate `parsed.Type` with `isValidSubjectToken` before interpolation. But the Platform service should also use a dot-free type: `"platform_tenant_created"` instead of `"platform.tenant.created"`.

**Fix in platform.go:**
```go
func (p *Platform) publishLifecycleEvent(tenantSlug, eventType string, data any) {
    if p.publisher == nil { return }
    if err := p.publisher.Notify(tenantSlug, map[string]any{
        "type": "platform_" + eventType, // no dots
        "data": data,
    }); err != nil {
        slog.Warn("publish lifecycle event failed", "event", eventType, "error", err)
    }
}
```

And change the caller: `p.publishLifecycleEvent(arg.Slug, "tenant_created", detail)`.

**Fix in pkg/nats/publisher.go (separate PR is fine):** Add `isValidSubjectToken(parsed.Type)` check after line 53.

---

## Debe corregirse

### C1. ExtractorConsumer: tree generation failure still marks document "ready"
**File:** `/home/enzo/rag-saldivia/services/ingest/internal/service/extractor_consumer.go:149-152`

When `c.treeGen.Generate()` fails (line 151), the error is logged but execution falls through to line 165 where the document is marked "ready". A document without a tree is not usable for search.

**Fix:** On tree generation failure, either:
- Set status to `"error"` with the tree error message, or
- Set a more granular status like `"pages_only"` (requires schema update), or
- Nak the message for retry (tree gen may be an LLM timeout)

Recommended: Nak with a log, let MaxDeliver=3 handle exhaustion, then on final delivery mark as "error".

### C2. ExtractorConsumer: Ack on unmarshal error loses the message permanently
**File:** `/home/enzo/rag-saldivia/services/ingest/internal/service/extractor_consumer.go:100-104`

Invalid JSON is Ack'd. While retrying bad JSON is pointless (it will never parse), the proper semantic is `msg.Term()` (terminally reject) rather than `msg.Ack()` (successfully processed). Using `Term()` correctly reflects the outcome in JetStream metrics and dead-letter monitoring.

**Fix:** Change `msg.Ack()` to `msg.Term()` on line 103.

### C3. ExtractorConsumer: not wired in main.go
**File:** `/home/enzo/rag-saldivia/services/ingest/cmd/main.go`

`NewExtractorConsumer` is defined but never instantiated or started in the service's main function. This means the consumer is dead code until wired. If this PR is meant to complete P0 wiring gaps, this should be wired.

**Fix:** Add to `cmd/main.go` after the worker start:
```go
extConsumer := service.NewExtractorConsumer(nc, resolver, treeGen)
if err := extConsumer.Start(ctx); err != nil {
    slog.Error("failed to start extractor consumer", "error", err)
    os.Exit(1)
}
defer extConsumer.Stop()
```

(Requires `tree.Generator` and `tenant.Resolver` to be instantiated first.)

### C4. TraceEvent eventType not validated for NATS subject safety
**File:** `/home/enzo/rag-saldivia/services/agent/internal/service/traces.go:62-75`

`TraceEvent` validates `tenantSlug` but NOT `eventType`. Currently `eventType` is not interpolated into the subject (it goes into the JSON payload only), so this is safe today. But the parameter name and doc comment (`llm_call, tool_call, error, etc.`) suggest it could be used in subjects in the future. This is a correctness note, not a blocker.

### C5. Platform service: UpdateTenant and DisableTenant/EnableTenant don't publish lifecycle events
**File:** `/home/enzo/rag-saldivia/services/platform/internal/service/platform.go:180-200`

Only `CreateTenant` publishes a lifecycle event. `UpdateTenant`, `DisableTenant`, and `EnableTenant` do not. If other services cache tenant state (e.g., `pkg/tenant.Resolver` caches pools), they won't know when a tenant is disabled.

**Fix:** Add lifecycle events:
```go
p.publishLifecycleEvent(slug, "tenant_updated", ...)
p.publishLifecycleEvent(slug, "tenant_disabled", ...)
p.publishLifecycleEvent(slug, "tenant_enabled", ...)
```

### C6. Platform cmd/main.go: NATS failure is a warn, not a fatal
**File:** `/home/enzo/rag-saldivia/services/platform/cmd/main.go:64-68`

NATS connection failure logs a warning and continues with `nc == nil`. This means `natspub.New(nil)` is called. Looking at `publisher.go:32`, `New(nil)` returns a `Publisher{nc: nil}`, and `Notify` will panic on nil dereference at line 58 (`p.nc.Publish`). There's no nil guard in `Notify`.

**Fix:** Either:
1. Add a nil guard in `Notify`: `if p.nc == nil { return fmt.Errorf("nats not connected") }`, or
2. Make NATS connection a fatal error for platform (since lifecycle events are now a feature)

Option 1 is safer (graceful degradation). The Platform service already handles the `publishLifecycleEvent` error path with a `slog.Warn`, so a nil publisher is fine as long as it doesn't panic.

---

## Sugerencias

### S1. ExtractorConsumer: add document_id validation
The `result.DocumentID` from the NATS message is used directly in SQL queries without validating it exists or is a valid UUID. A malicious or buggy extractor could target arbitrary document IDs. Consider a `GetDocument` check first, or at minimum validate UUID format.

### S2. PublishFeedback: "error_report" fires for non-errors
**File:** `/home/enzo/rag-saldivia/services/agent/internal/service/agent.go:284-289`

`pending_confirmation` triggers an `error_report` feedback event. A pending confirmation is not an error -- it's a normal flow pause. Consider:
```go
if status != "completed" && status != "pending_confirmation" {
```

### S3. File name sanitization: consider stripping non-ASCII and control characters
**File:** `/home/enzo/rag-saldivia/services/ingest/internal/handler/ingest.go:95-99`

`filepath.Base()` handles path traversal, but filenames with null bytes, Unicode RTL override (U+202E), or other control characters could cause issues in logging or filesystem operations. Consider a stricter sanitizer that replaces non-printable characters.

### S4. ExtractorConsumer: consider using Fetch batch instead of Next
The consumer uses `cons.Next()` (single message polling) in a tight loop. For throughput, `cons.Fetch(batchSize, ...)` would be more efficient, matching the pattern used in the ingest worker.

---

## Lo que esta bien

1. **traces.go PublishFeedback**: Properly validates both `tenantSlug` and `category` with `validateToken()` before NATS subject interpolation. Good defense in depth.

2. **traces.go safeToken regex**: `^[a-zA-Z0-9_-]+$` is strict and correct for NATS subject tokens. Rejects dots, spaces, wildcards.

3. **agent.go publishTraceEnd**: Clean separation -- trace events and feedback events fire from a single point, reducing the chance of missing a trace end.

4. **agent.go filterHistory**: Correctly rejects system/tool roles from user-supplied history (B1 from PR #71 review is fixed).

5. **agent.go output guardrails**: System prompt leak detection via `guardrails.ValidateOutput` is properly wired with the actual system prompt fragments.

6. **ingest handler**: `MaxBytesReader`, extension allowlist, `filepath.Base()` sanitization, and `writeJSON` helper all follow conventions.

7. **Platform cmd/main.go**: Proper NATS connection with `MaxReconnects(-1)` and `ReconnectWait(2s)`. Graceful drain on shutdown. OTEL setup with fallback.

8. **Platform service**: `slugRegex` validation on `CreateTenant`, duplicate key detection via PG error code, safe `TenantDetail` struct that strips connection strings.
