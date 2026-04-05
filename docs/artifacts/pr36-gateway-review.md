# Gateway Review — PR #36 feat/dockerfiles-ingest-service

**Fecha:** 2026-04-02
**Tipo:** review
**PR:** #36 (feat/dockerfiles-ingest-service)
**Intensity:** thorough
**Reviewer:** gateway-reviewer (Opus)

## Archivos revisados

| Archivo | Cambio |
|---|---|
| `services/chat/Dockerfile` | New |
| `services/rag/Dockerfile` | New |
| `services/ws/Dockerfile` | New |
| `services/ingest/Dockerfile` | New |
| `services/ingest/cmd/main.go` | New — entrypoint |
| `services/ingest/internal/handler/ingest.go` | New — HTTP handlers |
| `services/ingest/internal/service/ingest.go` | New — business logic |
| `services/ingest/internal/service/worker.go` | New — JetStream consumer |
| `services/ingest/db/migrations/001_init.up.sql` | New — schema |
| `services/ingest/go.mod` | New |
| `services/ingest/README.md` | Modified |
| `go.work` | Modified — added `./services/ingest` |

## Resultado

**CAMBIOS REQUERIDOS** (3 bloqueantes, 5 debe corregirse, 7 sugerencias)

---

## Hallazgos

### Bloqueantes

**B1. [ingest/cmd/main.go] No auth middleware -- all routes are unauthenticated**

The ingest service does not use `pkg/middleware.Auth()`. Every other service that exposes tenant-facing endpoints (chat, platform) either uses the shared auth middleware or verifies JWT explicitly in handlers. The ingest handler reads `X-User-ID` and `X-Tenant-Slug` from headers (lines 42-45 of handler), but these are trivially spoofable without JWT verification upstream.

In the current architecture, Traefik routes to services and each service verifies JWT locally (bible: "JWT verificacion local en cada servicio"). Without the auth middleware, anyone with network access to port 8007 can upload arbitrary files to any tenant's collection and read/delete any user's jobs.

**Fix:** Add `pkg/middleware.Auth(jwtSecret)` to the router in `cmd/main.go`, same pattern as other services. This requires adding a `JWT_SECRET` env var (which is already required by every other service). The auth middleware strips spoofed identity headers and sets them from verified JWT claims, which is exactly what the handler needs.

```go
import sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"

// in main():
jwtSecret := env("JWT_SECRET", "")
if jwtSecret == "" {
    slog.Error("JWT_SECRET is required")
    os.Exit(1)
}

r.Use(sdamw.Auth(jwtSecret))
```

**Severity:** Bloqueante. Bible: "La seguridad no es un tradeoff. Es una restriccion."

---

**B2. [ingest/db/migrations/001_init.up.sql:4] Type mismatch: UUID vs TEXT on FK to users(id)**

The `ingest_jobs` table defines:
```sql
id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
user_id  UUID NOT NULL REFERENCES users(id),
```

But the `users` table (from `services/auth/db/migrations/001_init.up.sql`) uses `TEXT` for its primary key:
```sql
id  TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
```

PostgreSQL will reject the FK constraint because the types don't match (`UUID` referencing `TEXT`). This migration will fail on any tenant DB that has the auth schema applied first (which is all of them -- auth is the base migration).

Additionally, the chat service uses `TEXT` for all its PKs and FKs, establishing a codebase convention of `TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text`.

**Fix:** Change both columns to `TEXT` to match the codebase convention:

```sql
id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
user_id     TEXT NOT NULL REFERENCES users(id),
```

Apply the same fix to the `connectors` table (`id` and `created_by` columns).

**Severity:** Bloqueante. The migration will fail at runtime. No workaround.

---

**B3. [ingest/internal/service/ingest.go:113] NATS payload built with fmt.Sprintf -- JSON injection**

```go
payload := fmt.Sprintf(`{"job_id":"%s","tenant_slug":"%s","collection":"%s","file_name":"%s","staged_path":"%s"}`,
    job.ID, tenantSlug, collection, fileName, stagedPath)
```

The `fileName` comes directly from `header.Filename` (multipart upload), which is user-controlled. A filename like `foo","staged_path":"/etc/passwd","extra":"bar` would break the JSON structure, potentially causing the worker to open an arbitrary file path. The `collection` field is also user-controlled.

This is a JSON injection that directly feeds into `im.StagedPath` in the worker, which is passed to `os.Open()`.

**Fix:** Use `json.Marshal` to produce the NATS payload:

```go
payload, err := json.Marshal(ingestMessage{
    JobID:      job.ID,
    TenantSlug: tenantSlug,
    Collection: collection,
    FileName:   fileName,
    StagedPath: stagedPath,
})
if err != nil {
    os.Remove(stagedPath)
    return nil, fmt.Errorf("marshal ingest message: %w", err)
}
```

This requires moving the `ingestMessage` struct from `worker.go` to a shared location in the package (or duplicating it in `ingest.go`). Since both files are in the same package (`service`), the struct is already accessible.

**Severity:** Bloqueante. Path traversal via crafted filename.

---

### Debe corregirse

**D1. [ingest/internal/handler/ingest.go:80-101] ListJobs filters by userID only -- missing tenant isolation**

`ListJobs` queries `WHERE user_id = $1`. In a per-tenant DB architecture this is safe because each tenant has its own PostgreSQL instance. However, the handler does NOT check `X-Tenant-Slug` at all:

```go
func (h *Ingest) ListJobs(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    if userID == "" {
        // only checks userID, not tenantSlug
    }
```

Same issue in `GetJob` and `DeleteJob` (lines 105, 128). For consistency with `Upload` (which checks both) and for defense-in-depth, all handlers should validate that `X-Tenant-Slug` is present. Once B1 is fixed (auth middleware), this header comes from verified JWT claims and provides an extra layer of tenant identity verification.

**Fix:** Add `tenantSlug` check to all handlers, consistent with `Upload`:

```go
userID := r.Header.Get("X-User-ID")
tenantSlug := r.Header.Get("X-Tenant-Slug")
if userID == "" || tenantSlug == "" {
    writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing identity headers"})
    return
}
```

---

**D2. [ingest/internal/handler/ingest.go:115] Error comparison via string matching is fragile**

```go
if err.Error() == "job not found" {
```

This pattern appears twice (lines 115 and 139). The chat service uses sentinel errors with `errors.Is()`:

```go
// chat pattern:
if errors.Is(err, service.ErrSessionNotFound) {
```

String comparison breaks silently if someone changes the error message or wraps the error.

**Fix:** Define sentinel errors in `service/ingest.go`:

```go
var ErrJobNotFound = errors.New("job not found")
```

Then use `errors.Is(err, service.ErrJobNotFound)` in the handler. Update `GetJob` and `DeleteJob` to return the sentinel error.

---

**D3. [ingest/internal/service/worker.go:201-202] Blueprint error response body leaked into error message**

```go
respBody, _ := io.ReadAll(resp.Body)
return fmt.Errorf("blueprint returned %d: %s", resp.StatusCode, string(respBody))
```

This reads the entire Blueprint error body and includes it in the error string. This error propagates to:
1. `processJob` line 144 where it's stored as `errMsg` in the DB via `UpdateJobStatus`
2. The `Job.Error` field which is returned to clients via `GetJob` (line 113 of handler)

The Blueprint response could contain internal infrastructure details, model config, or stack traces. These should not be exposed to end users.

**Fix:** Log the full response body server-side, but truncate and sanitize what goes into the DB error:

```go
respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
slog.Error("blueprint error", "status", resp.StatusCode, "body", string(respBody), "job_id", im.JobID)
return fmt.Errorf("blueprint returned HTTP %d", resp.StatusCode)
```

---

**D4. [ingest/internal/service/worker.go:211-213] Notification event uses job_id where user_id is expected**

```go
evt := map[string]string{
    "type":    "ingest.completed",
    "user_id": im.JobID, // worker doesn't have user_id; notification service looks it up
```

The comment acknowledges the problem but the fix is wrong. The `pkg/nats.Event` struct specifies `UserID` as "target user (who gets notified)". Sending the job ID in the `user_id` field means the notification service will try to look up a user by the job UUID and fail (or worse, silently drop the notification).

**Fix:** Include `user_id` in the NATS `ingestMessage` so the worker has it. The `Submit` function already has the `userID` parameter -- add it to the payload:

```go
type ingestMessage struct {
    JobID      string `json:"job_id"`
    TenantSlug string `json:"tenant_slug"`
    UserID     string `json:"user_id"`      // add this
    Collection string `json:"collection"`
    FileName   string `json:"file_name"`
    StagedPath string `json:"staged_path"`
}
```

Then in `publishCompletion`: `"user_id": im.UserID`.

---

**D5. [ingest/internal/service/worker.go:148] Failed job Nak'd for retry but status already set to "failed"**

```go
w.svc.UpdateJobStatus(ctx, im.JobID, "failed", &errMsg)
msg.Nak() // retry
```

The job status is set to `"failed"` in the DB, then the message is Nak'd for retry. When the message is redelivered (up to MaxDeliver=3), the worker will process it again, update status to `"processing"`, and retry the Blueprint call. This means:

1. A client polling `GetJob` between retries sees alternating `failed` -> `processing` -> `failed` status, which is confusing.
2. If the retry succeeds, the client saw `"failed"` but then it magically becomes `"completed"`.

**Fix:** Use an intermediate status like `"retrying"` on Nak, or only set `"failed"` on the final attempt. The `msg.Metadata()` method provides delivery count:

```go
meta, _ := msg.Metadata()
if meta != nil && meta.NumDelivered >= 3 {
    w.svc.UpdateJobStatus(ctx, im.JobID, "failed", &errMsg)
    msg.Term() // no more retries
} else {
    w.svc.UpdateJobStatus(ctx, im.JobID, "processing", nil) // keep as processing
    msg.Nak()
}
```

---

### Sugerencias

**S1. [ingest/cmd/main.go:104] WriteTimeout too low for large uploads**

```go
ReadTimeout:  120 * time.Second, // large uploads
WriteTimeout: 30 * time.Second,
```

`ReadTimeout` is correctly set to 120s for large uploads, but `WriteTimeout` is 30s. When the upload completes, the handler calls `Submit()` which stages the file to disk, writes to the DB, and publishes to NATS. If staging a 100MB file to disk takes more than 30s (e.g., slow disk), the write will timeout mid-response. Consider bumping to 60s.

---

**S2. [ingest/internal/service/ingest.go:48] os.MkdirAll error silently ignored**

```go
os.MkdirAll(cfg.StagingDir, 0750)
```

If the staging directory can't be created (permissions, readonly filesystem), the service will start successfully but every upload will fail. Consider checking the error at startup.

---

**S3. [ingest/internal/service/worker.go:107] Fetch batch size of 1 limits throughput**

```go
batch, err := w.cons.Fetch(1, jetstream.FetchMaxWait(5e9))
```

Fetching one message at a time is correct for a first version (simple, no concurrency bugs), but for a pipeline that processes 100MB files, the sequential bottleneck is the Blueprint HTTP call (potentially 30-120s per document). Multiple concurrent workers would improve throughput.

Consider adding a configurable worker concurrency (e.g., `INGEST_WORKERS=3`) with a semaphore pattern, as a follow-up.

---

**S4. [ingest/internal/service/worker.go:239-246] tenantFromSubject is defined but never used**

```go
func tenantFromSubject(subject string) string {
```

This function is dead code. Either use it in `processJob` to validate that the NATS subject matches the message payload's `tenant_slug`, or remove it.

Using it for validation would add defense-in-depth:
```go
if subjectTenant := tenantFromSubject(msg.Subject()); subjectTenant != im.TenantSlug {
    slog.Warn("tenant mismatch", "subject", subjectTenant, "payload", im.TenantSlug)
    msg.Term()
    return
}
```

---

**S5. [ingest/internal/handler/ingest.go:50] ParseMultipartForm with MaxUploadSize allocates full memory**

```go
if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
```

The `maxMemory` parameter to `ParseMultipartForm` is 100MB. This means Go will attempt to buffer up to 100MB in RAM per concurrent upload. With 10 concurrent uploads, that is 1GB of RAM. Consider using a lower memory threshold (e.g., 10MB) -- files larger than the threshold are automatically streamed to temp files by Go's multipart reader.

```go
if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB in memory, rest to disk
```

---

**S6. [ingest/internal/handler/ingest.go:62-63] No file type validation**

The handler accepts any file type. The NVIDIA Blueprint /v1/documents endpoint supports specific formats (PDF, DOCX, TXT, etc.). Uploading unsupported formats will succeed at the handler level but fail at the Blueprint, wasting staging disk space and queue capacity.

Consider adding a content-type or extension allowlist:
```go
allowedExts := map[string]bool{".pdf": true, ".docx": true, ".txt": true, ".md": true, ".csv": true}
ext := strings.ToLower(filepath.Ext(header.Filename))
if !allowedExts[ext] {
    writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported file type"})
    return
}
```

---

**S7. [ingest/internal/service/ingest.go:116-118] NATS publish failure silently continues**

```go
if err := s.nc.Publish(subject, []byte(payload)); err != nil {
    slog.Warn("failed to publish ingest job to NATS, will retry on restart", "error", err, "job_id", job.ID)
}
```

The comment says "will retry on restart" but there is no retry-on-restart mechanism. If NATS is down when Submit is called, the job is created in the DB as `pending` but no message is ever published. The job sits in `pending` status forever.

Options:
1. Return an error to the client (fail the upload if NATS is down)
2. Add a startup reconciliation loop that scans for `pending` jobs older than N minutes and re-publishes them
3. Use JetStream publish (with ack) instead of core NATS publish

Option 1 is simplest. Option 2 is most robust. Either way, the current behavior creates orphaned jobs.

---

## Lo que esta bien

1. **Dockerfile pattern is consistent.** All four new Dockerfiles follow the exact same pattern as existing services (auth, notification, platform): multi-stage build, `golang:1.25-alpine` builder, `distroless/static-debian12` runtime, `nonroot:nonroot` user. Port assignments are correct and non-overlapping (chat:8003, rag:8004, ws:8002, ingest:8007).

2. **Graceful shutdown is correct.** Signal handling, context propagation, worker Stop(), server Shutdown() -- all follow the established pattern from auth/chat/notification.

3. **NATS connection pattern is consistent.** Same retry/reconnect options as auth, chat, and notification services. Same publisher initialization pattern.

4. **JetStream usage is solid.** Durable consumer with explicit ack policy, max delivery of 3, file storage, 24h retention. The stream creation uses `CreateOrUpdateStream` which is idempotent. Good for service restarts.

5. **Tenant namespace isolation in Blueprint.** Collection names are correctly namespaced: `im.TenantSlug + "-" + im.Collection` (worker.go:169). This matches the spec: "Cada tenant tiene sus propias colecciones Milvus (namespaceo: `{tenant_slug}-{collection_name}`)".

6. **Ownership verification in GetJob and DeleteJob.** Both verify `j.UserID != userID` before returning data, preventing users from accessing other users' jobs within the same tenant. DeleteJob uses `WHERE id = $1 AND user_id = $2` which is correct.

7. **SQL uses parameterized queries.** All database queries use `$1`, `$2` placeholders (pgx). No string interpolation in SQL.

8. **Migration schema is well-designed.** The CHECK constraint on status, the partial index on active statuses (`WHERE status IN ('pending', 'processing')`), and the composite index on `(user_id, created_at DESC)` show good schema design.

9. **README is excellent.** Architecture diagram, endpoint table, NATS subjects, env vars, database tables -- all documented. Matches bible requirement: "doc en el mismo PR que el codigo".

10. **Error handling in handlers does not leak internals.** Generic error messages like "internal error" and "failed to queue document" are returned to clients. Stack traces and detailed errors go to `slog.Error` with request_id correlation.

---

## Resumen de acciones

| ID | Severidad | Archivo | Fix |
|---|---|---|---|
| B1 | Bloqueante | `cmd/main.go` | Add `pkg/middleware.Auth(jwtSecret)` + require `JWT_SECRET` env var |
| B2 | Bloqueante | `001_init.up.sql` | Change `UUID` to `TEXT` for `id`, `user_id`, `created_by` columns |
| B3 | Bloqueante | `service/ingest.go:113` | Replace `fmt.Sprintf` with `json.Marshal` for NATS payload |
| D1 | Debe corregirse | `handler/ingest.go` | Add `X-Tenant-Slug` check to ListJobs, GetJob, DeleteJob |
| D2 | Debe corregirse | `handler/ingest.go` | Replace string error comparison with sentinel errors + `errors.Is()` |
| D3 | Debe corregirse | `service/worker.go:201` | Truncate Blueprint error body, don't expose to clients via DB |
| D4 | Debe corregirse | `service/worker.go:211` | Add `user_id` to ingestMessage, fix notification event |
| D5 | Debe corregirse | `service/worker.go:148` | Only set "failed" on final retry attempt, not on every Nak |
| S1 | Sugerencia | `cmd/main.go:104` | Bump WriteTimeout to 60s |
| S2 | Sugerencia | `service/ingest.go:48` | Check MkdirAll error at startup |
| S3 | Sugerencia | `service/worker.go:107` | Consider configurable worker concurrency (follow-up) |
| S4 | Sugerencia | `service/worker.go:239` | Remove dead code `tenantFromSubject` or use it for validation |
| S5 | Sugerencia | `handler/ingest.go:50` | Lower ParseMultipartForm memory to 10MB |
| S6 | Sugerencia | `handler/ingest.go:62` | Add file extension allowlist |
| S7 | Sugerencia | `service/ingest.go:116` | Handle NATS publish failure (fail upload or add reconciliation) |
